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

package tool

import (
	"crypto/tls"
	"ezBastion/pkg/confmanager"
	"fmt"
	"path"

	"ezBastion/cmd/ezb_srv/models"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

func Trace(l *models.EzbLogs, c *gin.Context) {

	go func() {
		ep, _ := c.Get("exPath")
		exPath := ep.(string)
		cnf, _ := c.Get("configuration")
		conf := cnf.(*confmanager.Configuration)
		fcert := path.Join(exPath, conf.TLS.PublicCert)
		key := path.Join(exPath, conf.TLS.PrivateKey)
		ca := path.Join(exPath, conf.EZBPKI.CaCert)

		var log models.EzbLogs
		cert, err := tls.LoadX509KeyPair(fcert, key)
		if err != nil {
			fmt.Println(err)
			return
		}
		EzbDB := fmt.Sprintf("https://%s:%d/", conf.EZBDB.NetworkPKI.FQDN, conf.EZBDB.NetworkPKI.Port)
		client := resty.New()
		client.SetRootCertificate(ca)
		client.SetCertificates(cert)
		if l.ID == 0 {
			_, err := client.R().
				SetBody(l).
				SetResult(&log).
				Post(EzbDB + "logs")
			if err != nil {
				fmt.Println(err)
			}
		} else {
			_, err := client.R().
				SetBody(l).
				SetResult(&log).
				Put(EzbDB + "logs")
			if err != nil {
				fmt.Println(err)
			}
			// c.Set("trace", log)
		}
	}()

}

func IncRequest(l *models.EzbWorkers, c *gin.Context) {

	go func() {
		ep, _ := c.Get("exPath")
		exPath := ep.(string)
		cnf, _ := c.Get("configuration")
		conf := cnf.(*confmanager.Configuration)
		fcert := path.Join(exPath, conf.TLS.PublicCert)
		key := path.Join(exPath, conf.TLS.PrivateKey)
		ca := path.Join(exPath, conf.EZBPKI.CaCert)

		var wks models.EzbWorkers
		cert, err := tls.LoadX509KeyPair(fcert, key)
		if err != nil {
			fmt.Println(err)
			return
		}
		EzbDB := fmt.Sprintf("https://%s:%d/", conf.EZBDB.NetworkPKI.FQDN, conf.EZBDB.NetworkPKI.Port)
		client := resty.New()
		client.SetRootCertificate(ca)
		client.SetCertificates(cert)
		if l.ID != 0 {
			_, err := client.R().
				SetResult(&wks).
				Put(fmt.Sprintf("%sworkers/inc/%d", EzbDB, l.ID))
			if err != nil {
				fmt.Println(err)
			}
			// c.Set("trace", log)
		}
	}()

}
