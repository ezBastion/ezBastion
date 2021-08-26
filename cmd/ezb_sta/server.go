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
	"ezBastion/cmd/ezb_srv/cache"
	"ezBastion/cmd/ezb_srv/cache/memory"
	"ezBastion/cmd/ezb_sta/ctrl"
	"ezBastion/cmd/ezb_sta/middleware"
	"ezBastion/pkg/logmanager"
	"github.com/gin-gonic/gin"
	"path"
	"strconv"
)

// Must implement Mainservice interface from servicemanager package
type mainService struct{}

var storage cache.Storage

func (sm mainService) StartMainService(serverchan *chan bool) {
	logmanager.Debug("#### Main service started #####")
	// Pushing current conf to controllers
	server := gin.Default()
	storage = memory.NewStorage()

	server.Use(func(c *gin.Context) {
		c.Set("configuration", conf)
		c.Set("exPath", exePath)
	})

	server.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	})

	server.OPTIONS("*a", func(c *gin.Context) {
		c.AbortWithStatus(200)
	})
	// Init the caching system

	// Middleware
	server.Use(middleware.EzbAuthCacheJWT(storage, &conf, exePath))
	server.Use(middleware.EzbAuthJWT(storage, &conf, exePath))
	server.Use(middleware.EzbAuthForm)
	server.Use(middleware.SspiHandler())
	server.Use(middleware.EzbAuthSSPI)
	// token endpoint
	//route.POST("/token", middleware.EzbCache)
	server.POST("/token", ctrl.Createtoken(storage))
	server.GET("/token", ctrl.Createtoken(storage))
	server.GET("/renew", ctrl.Renewtoken(storage))
	server.GET("/access", ctrl.Renewtoken(storage))
	server.RunTLS(":"+strconv.Itoa(conf.EZBSTA.Network.Port), path.Join(exePath, conf.TLS.PublicCert), path.Join(exePath, conf.TLS.PrivateKey))
}
