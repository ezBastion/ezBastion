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
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"

	"ezBastion/cmd/ezb_wks/Middleware"
	"ezBastion/cmd/ezb_wks/models/exec"
	"ezBastion/cmd/ezb_wks/models/healthCheck"
	"ezBastion/cmd/ezb_wks/models/tasks"
	"ezBastion/cmd/ezb_wks/models/wkslog"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Must implement Mainservice interface from servicemanager package
type mainService struct{}

func (sm mainService) StartMainService(serverchan *chan bool) {

	log.Debug("start main")
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))
	r.Use(Middleware.ConfigMiddleware(conf))
	r.Use(Middleware.Limit)

	healthCheck.Routes(r)
	wkslog.Routes(r)
	exec.Routes(r)
	tasks.Routes(r)
	caCert, err := ioutil.ReadFile(path.Join(exePath, conf.EZBPKI.CaCert))
	if err != nil {
		log.Fatal(err)

	}
	listen := fmt.Sprintf("%s:%d", conf.EZBWKS.Network.FQDN, conf.EZBWKS.Network.Port)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfigPKI := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
		MinVersion: tls.VersionTLS12,
	}
	tlsConfigPKI.BuildNameToCertificate()
	srv := &http.Server{
		Addr:      listen,
		TLSConfig: tlsConfigPKI,
		Handler:   r,
	}

	go func() {
		if err := srv.ListenAndServeTLS(path.Join(exePath, conf.TLS.PublicCert), path.Join(exePath, conf.TLS.PrivateKey)); err != nil {
			log.Debug("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Debug("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Debug("Server exiting")
}
