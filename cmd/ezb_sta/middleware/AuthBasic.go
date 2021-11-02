package middleware

import (
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jtblin/go-ldap-client"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func EzbAuthBasic(ldapclient *ldap.LDAPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		logg := log.WithFields(log.Fields{"Middleware": "basic"})
		auth := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || strings.ToLower(auth[0]) != "basic" {
			return
		}

		if strings.ToLower(auth[0]) == "basic" {
			payload, _ := base64.StdEncoding.DecodeString(auth[1])
			pair := strings.SplitN(string(payload), ":", 2)

			if len(pair) != 2 {
				logg.Error("bad request #BSC0001: ")
				c.AbortWithError(http.StatusBadRequest, errors.New("#STA-BSC0001"))
				return
			}

			// check the user in the database
			username := pair[0]
			password := pair[1]

			err := CheckUserandSet(c, username, password, ldapclient)
			if err != "" {
				logg.Error("error thrown when checking user " + err)
			}
		}
	}
}
