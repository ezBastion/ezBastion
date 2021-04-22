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

package certmanager

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path"
	"time"
)

func NewRootCertificate(caCert, caKey string, addresses []string) error {
	keyOut, _ := os.OpenFile(caKey, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		panic(err)
	}

	err = pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	if err != nil {
		return err
	}
	err = keyOut.Close()
	if err != nil {
		return err
	}

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Organization: []string{"ezBastion"},
			CommonName:   "ezBastion PKI",
		},
		DNSNames:              addresses,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(20, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.ECDSAWithSHA256,
	}
	pub := &priv.PublicKey
	caB, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(caCert)
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: caB})
	if err != nil {
		return err
	}
	err = certOut.Close()
	if err != nil {
		return err
	}

	return nil
}

func NewCertificateRequest(commonName string, addresses []string) *x509.CertificateRequest {
	serial := uuid.NewV4()

	certificate := x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"ezBastion"},
			CommonName:   commonName,
			SerialNumber: serial.String(),
		},

		SignatureAlgorithm: x509.ECDSAWithSHA256,
	}

	for i := 0; i < len(addresses); i++ {
		if ip := net.ParseIP(addresses[i]); ip != nil {
			certificate.IPAddresses = append(certificate.IPAddresses, ip)
		} else {
			certificate.DNSNames = append(certificate.DNSNames, addresses[i])
		}
	}

	return &certificate
}

func Generate(certificate *x509.CertificateRequest, ezbpki, certFilename, keyFilename, caFileName, exePath string) error {

	var priv *ecdsa.PrivateKey
	caCSR := path.Join(exePath, "cert/ezbastion.csr")
	var derBytes []byte

	//did we have a private key ?
	if _, err := os.Stat(keyFilename); os.IsNotExist(err) {
		// generate private key
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return fmt.Errorf("failed to generate private key: %v", err)
		}
		// open new key file
		keyOut, err := os.OpenFile(keyFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return fmt.Errorf("failed to open key %v for writing: %v", keyFilename, err)
		}
		// convert EC to ASN.1
		b, err := x509.MarshalECPrivateKey(priv)
		if err != nil {
			return fmt.Errorf("failed to marshal priv: %v", err)
		}
		// PEM encode to file
		err = pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
		if err != nil {
			return err
		}
		err = keyOut.Close()
		if err != nil {
			return err
		}
	} else {

		raw, err := ioutil.ReadFile(keyFilename)
		if err != nil {
			return err
		}
		var blockKey *pem.Block
		blockKey, _ = pem.Decode(raw)
		if blockKey == nil || blockKey.Type != "EC PRIVATE KEY" {
			return fmt.Errorf("failed to decode PEM block containing private key")
		}
		priv, err = x509.ParseECPrivateKey(blockKey.Bytes)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(caCSR); os.IsNotExist(err) {
		//fmt.Println("file", caCSR, "not found")
		derBytes, err = x509.CreateCertificateRequest(rand.Reader, certificate, priv)
		if err != nil {
			return err
		}
	} else {
		var block *pem.Block
		raw, readerror := ioutil.ReadFile(caCSR)
		if readerror != nil {
			return err
		}
		block, _ = pem.Decode(raw)
		if block == nil || block.Type != "CERTIFICATE REQUEST" {
			return fmt.Errorf("failed to decode PEM block containing public key")
		}
		pcr, err := x509.ParseCertificateRequest(block.Bytes)
		if err != nil {
			return err
		}
		err = pcr.CheckSignature()
		if err != nil {
			return err
		}
		derBytes = pcr.Raw
	}

	//fmt.Println("Created Certificate Signing Request for client.")
	conn, err := net.Dial("tcp", ezbpki)
	if err != nil {
		return err
	}
	defer conn.Close()
	//fmt.Println("Successfully connected to Root Certificate Authority.")
	writer := bufio.NewWriter(conn)
	// Send two-byte header containing the number of ASN1 bytes transmitted.
	header := make([]byte, 2)
	binary.LittleEndian.PutUint16(header, uint16(len(derBytes)))
	_, err = writer.Write(header)
	if err != nil {
		return err
	}
	// Now send the certificate request data
	_, err = writer.Write(derBytes)
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	//fmt.Println("Transmitted Certificate Signing Request to RootCA.")
	// The RootCA will now send our signed certificate back for us to read.
	reader := bufio.NewReader(conn)
	// Read header containing the size of the ASN1 data.
	certHeader := make([]byte, 2)
	_, err = reader.Read(certHeader)
	if err == io.EOF {
		//save CSR
		csrOut, _ := os.Create(caCSR)
		err := pem.Encode(csrOut, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: derBytes})
		if err != nil {
			return err
		}
		err = csrOut.Close()
		if err != nil {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}
	certSize := binary.LittleEndian.Uint16(certHeader)
	// Now read the certificate data.
	certBytes := make([]byte, certSize)
	_, err = reader.Read(certBytes)
	if err != nil {
		return err
	}
	newCert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return err
	}

	// Finally, the RootCA will send its own certificate back so that we can validate the new certificate.
	rootCertHeader := make([]byte, 2)
	_, err = reader.Read(rootCertHeader)
	if err != nil {
		return err
	}
	rootCertSize := binary.LittleEndian.Uint16(rootCertHeader)
	// Now read the certificate data.
	rootCertBytes := make([]byte, rootCertSize)
	_, err = reader.Read(rootCertBytes)
	if err != nil {
		return err
	}
	rootCert, err := x509.ParseCertificate(rootCertBytes)
	if err != nil {
		return err
	}

	err = ValidateCertificate(newCert, rootCert)
	if err != nil {
		return err
	}
	// all good save the files

	certOut, err := os.Create(certFilename)
	if err != nil {
		return fmt.Errorf("failed to open %v for writing: %v", certFilename, err)
	}
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return err
	}
	err = certOut.Close()
	if err != nil {
		return err
	}

	if _, err := os.Stat(caFileName); os.IsNotExist(err) {
		caOut, err := os.Create(caFileName)
		if err != nil {
			return fmt.Errorf("failed to open %v for writing: %v", caFileName, err)
		}
		err = pem.Encode(caOut, &pem.Block{Type: "CERTIFICATE", Bytes: rootCertBytes})
		if err != nil {
			return err
		}
		err = caOut.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func ValidateCertificate(newCert *x509.Certificate, rootCert *x509.Certificate) error {
	// how to check key with openssl: 3 command below must have the same pubkey
	//openssl req  -pubkey -in ezbastion.csr
	//openssl x509 -pubkey -in ezbastion.crt
	//openssl ec   -pubout -in ezbastion.key

	roots := x509.NewCertPool()
	roots.AddCert(rootCert)
	verifyOptions := x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	_, err := newCert.Verify(verifyOptions)
	if err != nil {
		return err
	}
	return nil
}
