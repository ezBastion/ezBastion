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
	"os/signal"
	"strings"
	"time"

	"ezBastion/cmd/ezb_srv/cache"
	"ezBastion/cmd/ezb_srv/cache/memory"
	"ezBastion/cmd/ezb_srv/ctrl"
	"ezBastion/cmd/ezb_srv/middleware"
	"ezBastion/cmd/ezb_srv/models"
	"github.com/gin-contrib/location"

	"net/http"
	"os"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var storage cache.Storage
// Must implement Mainservice interface from servicemanager package
type mainService struct{}

func (sm mainService) StartMainService(serverchan *chan bool)  {

	storage = memory.NewStorage()
	listen := fmt.Sprintf("%s:%d", conf.EZBSRV.Network.FQDN,conf.EZBSRV.Network.Port)

	/* gin */
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))
	r.Use(middleware.AddHeaders)
	r.OPTIONS("*a", func(c *gin.Context) {
		c.AbortWithStatus(200)
	})
	r.Use(middleware.LoadConfig(&conf, exePath))
	r.Use(middleware.StartTrace)
	r.Use(middleware.InternalWork(storage, &conf))
	r.Use(middleware.AuthJWT(storage, &conf, exePath))
	r.Use(middleware.Store(storage, &conf))
	r.Use(middleware.RouteParser)
	r.Use(middleware.GetParams(storage, &conf))
	r.Use(middleware.SelectWorker)
	r.Use(location.Default())
	r.GET("*a", sendAction)
	r.POST("*a", sendAction)
	r.PUT("*a", sendAction)
	r.DELETE("*a", sendAction)
	r.PATCH("*a", sendAction)

	// log.Fatal(http.ListenAndServe(conf.Listen, r))
	srv := &http.Server{
		Addr:    listen,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Debug("listen: %s\n", err)
		}
	}()
	/* gin */

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

func sendAction(c *gin.Context) {

	escapedPath := c.Request.URL.EscapedPath()
	path := strings.Split(escapedPath, "/")
	routeType := c.MustGet("routeType").(string)
	log.Debug("sendAction routeType: ", routeType)
	switch routeType {
	case "worker":
		ctrl.SendAction(c, storage)
		break
	case "internal":
		account := c.MustGet("account").(models.EzbAccounts)
		if account.Name == "anonymous" {
			c.JSON(401, "MUST LOGIN")
			return
		}

		action := path[3]
		cmd := path[4]
		log.Debug("sendAction internal  action:", action, " cmd:", cmd)
		switch action {
		case "log":
			switch cmd {
			case "xtrack":
				ctrl.GetXtrack(c)
				break
			case "last":
				ctrl.GetLog(c)
				break
			}
			break
		case "healthcheck":
			switch cmd {
			case "jobs":
				ctrl.GetJobs(c)
				break
			case "load":
				ctrl.GetLoad(c)
				break
			case "scripts":
				ctrl.GetScripts(c)
				break
			case "conf":
				ctrl.GetConf(c)
				break
			}
			break
		}
		break
	case "tasks":
		ctrl.GetTask(c)
		break
	}
}
