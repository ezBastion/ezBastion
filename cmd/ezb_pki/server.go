// This file is part of ezBastion.

//     ezBastion is free software: you can redistribute it and/or modify
//     it under the terms of the GNU Affero General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.

//     ezBastion is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU Affero General Public License for more details.

//     You should have received a copy of the GNU Affero General Public License
//     along with ezBastion.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
)

// Must implement Mainservice interface from servicemanager package
type mainService struct{}

func (sm mainService) StartMainService(serverchan *chan bool) {
	caPath := path.Join(exePath, conf.EZBPKI.CaCert)
	keyPath := path.Join(exePath, conf.EZBPKI.CaKey)

	caPublicKeyFile, err := ioutil.ReadFile(caPath)
	if err != nil {
		log.Fatal(err)
	}
	pemBlock, _ := pem.Decode(caPublicKeyFile)

	if pemBlock == nil || pemBlock.Type != "CERTIFICATE" {
		log.Fatal("failed to decode PEM block containing public key")
	}
	caCRT, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Root CA loaded.")

	fp := sha1.Sum(caCRT.Raw)
	log.Debug("fingerprint, %v\n ", fp)

	caPrivateKeyFile, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatal(err)
	}
	pemBlock, _ = pem.Decode(caPrivateKeyFile)
	if pemBlock == nil || pemBlock.Type != "EC PRIVATE KEY" {
		log.Fatal("failed to decode PEM block containing private key")
	}
	log.Debug("Private key Decode.")
	caPrivateKey, err := x509.ParseECPrivateKey(pemBlock.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Private key loaded.")

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", conf.EZBPKI.Network.FQDN, conf.EZBPKI.Network.Port))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listen at ", fmt.Sprintf("%s:%d", conf.EZBPKI.Network.FQDN, conf.EZBPKI.Network.Port))
	defer func() {
		listener.Close()
		log.Debug("Listener closed")
	}()

	for {

		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
			break
		}
		go signconn(conn, caCRT, caPrivateKey)
	}
}

func signconn(conn net.Conn, rootCert *x509.Certificate, privateKey *ecdsa.PrivateKey) error {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	header := make([]byte, 2)
	_, err := reader.Read(header)
	if err != nil {
		log.Println(err)
		return err
	}
	asn1DataSize := binary.LittleEndian.Uint16(header)

	asn1Data := make([]byte, asn1DataSize)
	_, err = reader.Read(asn1Data)
	if err != nil {
		log.Println(err)
		return err
	}
	clientCSR, err := x509.ParseCertificateRequest(asn1Data)
	if err != nil {
		log.Println(err)
		return err
	}
	if err = clientCSR.CheckSignature(); err != nil {
		log.Println(err)
		return err
	}
	clientCRTTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Signature:             clientCSR.Signature,
		SignatureAlgorithm:    clientCSR.SignatureAlgorithm,
		PublicKey:             clientCSR.PublicKey,
		PublicKeyAlgorithm:    clientCSR.PublicKeyAlgorithm,
		Issuer:                rootCert.Subject,
		Subject:               clientCSR.Subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		DNSNames:              clientCSR.DNSNames,
		BasicConstraintsValid: true,
	}
	certData, err := x509.CreateCertificate(rand.Reader, clientCRTTemplate, rootCert, clientCSR.PublicKey, privateKey)
	if err != nil {
		log.Println(err)
		return err
	}

	writer := bufio.NewWriter(conn)

	certHeader := make([]byte, 2)
	binary.LittleEndian.PutUint16(certHeader, uint16(len(certData)))
	_, err = writer.Write(certHeader)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = writer.Write(certData)
	if err != nil {
		log.Println(err)
		return err
	}

	rootCertHeader := make([]byte, 2)
	binary.LittleEndian.PutUint16(rootCertHeader, uint16(len(rootCert.Raw)))
	_, err = writer.Write(rootCertHeader)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = writer.Write(rootCert.Raw)
	if err != nil {
		log.Println(err)
		return err
	}

	err = writer.Flush()
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Transmitted client Certificate to ", clientCSR.Subject.CommonName)

	return nil
}
