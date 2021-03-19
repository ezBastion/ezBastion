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
	"math/big"
	"net"
	"os"
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

	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	keyOut.Close()

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
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: caB})
	certOut.Close()

	return nil
}

func NewCertificateRequest(commonName string, duration int, addresses []string) *x509.CertificateRequest {
	certificate := x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"ezBastion"},
			CommonName:   commonName,
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

func Generate(certificate *x509.CertificateRequest, ezbpki, certFilename, keyFilename, caFileName string) error {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("Failed to generate private key: %v", err)
	}

	derBytes, err := x509.CreateCertificateRequest(rand.Reader, certificate, priv)
	if err != nil {
		return err
	}
	fmt.Println("Created Certificate Signing Request for client.")
	conn, err := net.Dial("tcp", ezbpki)
	if err != nil {
		return err
	}
	defer conn.Close()
	fmt.Println("Successfully connected to Root Certificate Authority.")
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
	fmt.Println("Transmitted Certificate Signing Request to RootCA.")
	// The RootCA will now send our signed certificate back for us to read.
	reader := bufio.NewReader(conn)
	// Read header containing the size of the ASN1 data.
	certHeader := make([]byte, 2)
	_, err = reader.Read(certHeader)
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
	fmt.Println("Received new Certificate from RootCA.")
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
	fmt.Println("Received Root Certificate from RootCA.")
	rootCert, err := x509.ParseCertificate(rootCertBytes)
	if err != nil {
		return err
	}

	err = ValidateCertificate(newCert, rootCert)
	if err != nil {
		return err
	}
	// all good save the files
	keyOut, err := os.OpenFile(keyFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Failed to open key %v for writing: %v", keyFilename, err)
	}
	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return fmt.Errorf("Failed to marshal priv: %v", err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	keyOut.Close()

	certOut, err := os.Create(certFilename)
	if err != nil {
		return fmt.Errorf("Failed to open %v for writing: %v", certFilename, err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	certOut.Close()

	caOut, err := os.Create(caFileName)
	if err != nil {
		return fmt.Errorf("Failed to open %v for writing: %v", caFileName, err)
	}
	pem.Encode(caOut, &pem.Block{Type: "CERTIFICATE", Bytes: rootCertBytes})
	caOut.Close()
	return nil
}

func ValidateCertificate(newCert *x509.Certificate, rootCert *x509.Certificate) error {
	roots := x509.NewCertPool()
	roots.AddCert(rootCert)
	verifyOptions := x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	_, err := newCert.Verify(verifyOptions)
	if err != nil {
		fmt.Println("Failed to verify chain of trust.")
		return err
	}
	fmt.Println("Successfully verified chain of trust.")

	return nil
}
