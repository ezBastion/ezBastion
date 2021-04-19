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

package setupmanager

import (
	"bufio"
	"ezBastion/pkg/certmanager"
	"ezBastion/pkg/confmanager"
	"fmt"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"

	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func ExeFullPath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err

}

func ExePath() (string, error) {
	p, err := ExeFullPath()
	if err != nil {
		return "", err
	}
	return filepath.Dir(p), nil
}

//CheckFolder
func CheckFolder(exePath string, SERVICENAME string) error {
	folders := []string{"log", "cert", "conf"}
	switch SERVICENAME {
	case "ezb_db":
		{
			folders = append(folders, "db")
		}
	case "ezb_pki":
		{
			folders = append(folders, "db")
		}
	case "ezb_wks":
		{
			folders = append(folders, "script", "job")
		}
	}
	for _, folder := range folders {
		if _, err := os.Stat(path.Join(exePath, folder)); os.IsNotExist(err) {
			err = os.MkdirAll(path.Join(exePath, folder), 0600)
			if err != nil {
				return err
			}
			log.Println("Create ", folder, " folder.")
		}
	}
	return nil
}

//AskForConfirmation waiting for user yes or no
func AskForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("\n%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
func AskForValue(s, def string, pattern string) string {
	reader := bufio.NewReader(os.Stdin)
	re := regexp.MustCompile(pattern)
	for {
		fmt.Printf("%s [%s]: ", s, def)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}

		response = strings.TrimSpace(response)
		if response == "" {
			return def
		} else if re.MatchString(response) {
			return response
		} else {
			fmt.Printf("[%s] wrong format, must match (%s)\n", response, pattern)
		}
		fmt.Printf("[%s] wrong format, must match (%s)\n", response, pattern)
	}
}

func Setup(exePath, confPath, SERVICENAME string) error {
	err := CheckFolder(exePath, SERVICENAME)
	if err != nil {
		return err
	}
	conf, err := confmanager.CheckConfig(confPath, exePath)
	if err != nil {
		log.Errorf("Setup error: %v", err)
		c, _ := toml.Marshal(conf)
		wferr := ioutil.WriteFile(confPath, c, 0600)
		if wferr != nil {
			log.Fatal(wferr)
		}
		log.Println("New ", confPath, " file created, please fill it.")
		log.Println("Then, restart init to generate certificate.")
		return err
	}
	keyFile := path.Join(exePath, conf.TLS.PrivateKey)
	certFile := path.Join(exePath, conf.TLS.PublicCert)
	caCert := path.Join(exePath, conf.EZBPKI.CaCert)
	caKey := path.Join(exePath, conf.EZBPKI.CaKey)
	_, ficacert := os.Stat(caCert)
	_, ficakey := os.Stat(caKey)
	_, fipriv := os.Stat(keyFile)
	_, fipub := os.Stat(certFile)

	if SERVICENAME == "ezb_pki" {
		if os.IsNotExist(ficacert) || os.IsNotExist(ficakey) {
			err := certmanager.NewRootCertificate(caCert, caKey, conf.TLS.SAN)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Root certificate saved at ", caCert)
		}
		return nil
	} else if SERVICENAME == "ezb_setup" {
		return nil
	} else {
		if os.IsNotExist(ficacert) {
			log.Fatalln("PKI public certificate not found")
		}
		if os.IsNotExist(fipriv) || os.IsNotExist(fipub) {
			ezbPKI := fmt.Sprintf("%s:%d", conf.EZBPKI.Network.FQDN, conf.EZBPKI.Network.Port)
			conn, err := net.Dial("tcp", ezbPKI)
			if err != nil {
				log.Fatalln("Failed to connect PKI at %s ", ezbPKI)
			} else {
				conn.Close()
			}

			request := certmanager.NewCertificateRequest("ezBastion", 730, conf.TLS.SAN)
			certmanager.Generate(request, ezbPKI, certFile, keyFile, caCert, exePath)
		}
	}
	return nil
}
