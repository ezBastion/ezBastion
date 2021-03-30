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
	"fmt"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"
)




// Must implement Mainservice interface from servicemanager package
type mainService struct{}
func (sm mainService) StartMainService(serverchan *chan bool) {
	listen := fmt.Sprintf("%s:%d", conf.EZBADM.Network.FQDN,conf.EZBADM.Network.Port)
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))

	r.GET("/config.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ezb_db":fmt.Sprintf("https://%s:%d/", conf.EZBDB.NetworkJWT.FQDN,conf.EZBDB.NetworkJWT.Port),"ezb_srv": conf.EZBSRV.ExternalURL })
	} )
	r.StaticFile("/favicon.ico", path.Join(exePath, "docs","favicon.ico"))
	r.GET("/", func(c *gin.Context) {
	 c.File(	path.Join(exePath, "docs","index.html"))
	})
	r.Static("/app", path.Join(exePath, "docs","app"))
	r.Static("/assets", path.Join(exePath, "docs","assets"))
	r.Static("/bower_components", path.Join(exePath, "docs","bower_components"))
	r.Static("/scripts", path.Join(exePath, "docs","scripts"))
	r.Static("/styles", path.Join(exePath, "docs","styles"))

	srv := &http.Server{
		Addr:    listen,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
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

