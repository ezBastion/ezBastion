package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func EzbAuthCacheJWT(c *gin.Context) {

	logg := log.WithFields(log.Fields{"Middleware": "cache_jwt"})
	authHead := c.GetHeader("Authorization")

	bearer := strings.Split(authHead, " ")
	if len(bearer) != 2 {
		logg.Error("bad Authorization #CJ0001: " + authHead)
		c.AbortWithError(http.StatusForbidden, errors.New("#CSTA-JWT0001"))
		return
	}
	if strings.Compare(strings.ToLower(bearer[0]), "bearer") != 0 {
		if strings.Compare(strings.ToLower(bearer[0]), "negotiate") == 0 {
			return
		}
		return
	}
	// OK, we can check the cache to find the token
	//tokenString := bearer[1]

}
