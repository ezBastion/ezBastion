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

package setup

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"time"

	"ezBastion/cmd/ezb_pki/models"

	"ezBastion/pkg/setupmanager"

	"github.com/Showmax/go-fqdn"

	"github.com/urfave/cli"
)

var exPath string
var confFile string

func init() {
	ex, _ := os.Executable()
	exPath = filepath.Dir(ex)
	confFile = path.Join(exPath, "conf/config.json")
}

// CheckConfig test if config.json match the model
func CheckConfig() (conf models.Configuration, err error) {

	raw, err := ioutil.ReadFile(confFile)
	if err != nil {
		return conf, err
	}
	err = json.Unmarshal(raw, &conf)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

func Setup() error {
	fqdn := fqdn.Get()
	hostname, _ := os.Hostname()
	quiet := true
	err := setupmanager.CheckFolder(exPath)
	if err != nil {
		return err
	}
	conf, err := CheckConfig()
	if err != nil {
		quiet = false
		conf.Listen = "0.0.0.0:5010"
		conf.ServiceName = "ezb_pki"
		conf.ServiceFullName = "ezBastion PKI"
		conf.Logger.LogLevel = "warning"
		conf.Logger.MaxSize = 5
		conf.Logger.MaxBackups = 10
		conf.Logger.MaxAge = 180
	}
	if quiet == false {
		fmt.Println("\nWhich port do you want to listen to?")
		fmt.Println("ex: :5010, 0.0.0.0:5100, localhost:7800, name.domain:2000 ...")
		for {
			listen := setupmanager.AskForValue("listen", conf.Listen, "^[\\.0-9|\\w]*:[0-9]{1,5}$")
			c := setupmanager.AskForConfirmation(fmt.Sprintf("Listen on (%s) ok?", listen))
			if c {
				conf.Listen = listen
				break
			}
		}

		fmt.Println("\nWhat is service name?")
		fmt.Println("ex: ezb_pki, myPKI-p5010, api-pki-uat ...")
		for {
			name := setupmanager.AskForValue("name", conf.ServiceName, "^[\\w-]+$")
			c := setupmanager.AskForConfirmation(fmt.Sprintf("Service name (%s) ok?", name))
			if c {
				conf.ServiceName = name
				break
			}
		}

		fmt.Println("\nWhat is service full name?")
		fmt.Println("ex: my pki service, Api PKI for UAT ...")
		for {
			fullname := setupmanager.AskForValue("full name", conf.ServiceFullName, "^[\\w -]+$")
			c := setupmanager.AskForConfirmation(fmt.Sprintf("Service full name (%s) ok?", fullname))
			if c {
				conf.ServiceFullName = fullname
				break
			}
		}

		c, _ := json.Marshal(conf)
		ioutil.WriteFile(confFile, c, 0600)
		log.Println(confFile, " saved.")

	}

	keyfile := path.Join(exPath, "cert/"+conf.ServiceName+"-ca.key")
	if _, err := os.Stat(keyfile); os.IsNotExist(err) {
		keyOut, _ := os.OpenFile(keyfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

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
		log.Println("Private key saved at " + keyfile)

		ca := &x509.Certificate{
			SerialNumber: big.NewInt(1653),
			Subject: pkix.Name{
				Organization: []string{"ezBastion"},
				CommonName:   conf.ServiceName,
			},
			DNSNames:              []string{hostname, fqdn},
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
			return cli.NewExitError(err, -1)
		}

		rootCAfile := path.Join(exPath, "cert/"+conf.ServiceName+"-ca.crt")
		certOut, err := os.Create(rootCAfile)
		pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: caB})
		certOut.Close()
		log.Println("Root certificat saved at ", rootCAfile)
	}
	return nil
}
