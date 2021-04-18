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
	"ezBastion/pkg/confmanager"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"

	"ezBastion/cmd/ezb_db/configuration"
	"ezBastion/cmd/ezb_db/routes"
	"golang.org/x/sync/errgroup"

	"ezBastion/cmd/ezb_db/Middleware"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

var g errgroup.Group

func routerJWT(db *gorm.DB, lic configuration.License, conf confmanager.Configuration) http.Handler {
	loggerJWT := log.WithFields(log.Fields{"module": "jwt", "type": "http"})
	r := gin.Default()
	r.Use(ginrus.Ginrus(loggerJWT, time.RFC3339, true))
	r.Use(Middleware.AddHeaders)
	r.OPTIONS("*a", func(c *gin.Context) {
		c.AbortWithStatus(200)
	})
	r.Use(Middleware.DBMiddleware(db))
	r.Use(Middleware.AuthJWT(db, conf))
	r.Use(Middleware.LicenseMiddleware(lic))
	routes.Routes(r)
	return r
}

func routerPKI(db *gorm.DB, lic configuration.License) http.Handler {
	loggerPKI := log.WithFields(log.Fields{"module": "pki", "type": "http"})
	r := gin.Default()
	r.Use(ginrus.Ginrus(loggerPKI, time.RFC3339, true))
	r.Use(Middleware.AddHeaders)
	r.OPTIONS("*a", func(c *gin.Context) {
		c.AbortWithStatus(200)
	})
	r.Use(Middleware.DBMiddleware(db))
	r.Use(Middleware.LicenseMiddleware(lic))
	routes.Routes(r)
	return r
}

// Must implement Mainservice interface from servicemanager package
type mainService struct{}

func (sm mainService) StartMainService(serverchan *chan bool) {

	log.WithFields(log.Fields{"module": "main", "type": "log"})
	log.Debug("loglevel: ", conf.Logger.LogLevel)

	lic := configuration.License{}

	gin.SetMode(gin.ReleaseMode)
	db, err := configuration.InitDB(conf, exePath)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	err = configuration.InitLic(&lic, db)
	if err != nil {
		log.Fatal(err)
	}
	caCert, err := ioutil.ReadFile(path.Join(exePath, conf.EZBPKI.CaCert))
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	/* listner jwt */
	listenJWT := fmt.Sprintf("%s:%d", conf.EZBDB.NetworkJWT.FQDN, conf.EZBDB.NetworkJWT.Port)

	tlsConfigJWT := &tls.Config{}
	serverJWT := &http.Server{
		Addr:      listenJWT,
		TLSConfig: tlsConfigJWT,
		Handler:   routerJWT(db, lic, conf),
	}
	/* listner jwt */
	/* listner pki */

	listenPKI := fmt.Sprintf("%s:%d", conf.EZBDB.NetworkPKI.FQDN, conf.EZBDB.NetworkPKI.Port)

	tlsConfigPKI := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
		MinVersion: tls.VersionTLS12,
	}
	tlsConfigPKI.BuildNameToCertificate()
	serverPKI := &http.Server{
		Addr:      listenPKI,
		TLSConfig: tlsConfigPKI,
		Handler:   routerPKI(db, lic),
	}
	/* listner pki */

	g.Go(func() error {
		return serverJWT.ListenAndServeTLS(path.Join(exePath, conf.TLS.PublicCert), path.Join(exePath, conf.TLS.PrivateKey))
	})

	g.Go(func() error {
		return serverPKI.ListenAndServeTLS(path.Join(exePath, conf.TLS.PublicCert), path.Join(exePath, conf.TLS.PrivateKey))
	})
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = serverJWT.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
