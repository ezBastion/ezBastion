package middleware

import (
	"errors"
	"ezBastion/cmd/ezb_sta/models"
	"github.com/gin-gonic/gin"
	"github.com/quasoft/websspi"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

var (
	auth   *websspi.Authenticator
	config *websspi.Config
	h      http.Handler
)

func init() {
	config = websspi.NewConfig()
	auth, _ = websspi.New(config)
	auth.Config.AuthUserKey = "X-Authenticated-user"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Currently no needs to do something, websspi will try to set the userinfo
	})
	// try to use the handler to do the sspi
	h = auth.WithAuth(handler)
}

func EzbAuthSSPI(c *gin.Context) {

	// SSPI middleware changes result, so it must be set at the end, and exit immediately if one of the other middlerware
	// handled the context
	// Hnadle only AD requests
	a, err := c.Get("aud")
	if err {
		if a == "internal" {
			return
		}
	}
	// if request is jwt request, abort
	a, err = c.Get("jwt")
	if err {
		return
	}

	authHead := c.GetHeader("Authorization")
	username := c.GetHeader("X-Authenticated-user")
	logg := log.WithFields(log.Fields{"middleware": "sspi"})

	if authHead != "" {
		nego := strings.Split(authHead, " ")
		if len(nego) != 2 {
			logg.Error("bad Negotiation #SSPI0001: " + authHead)
			c.AbortWithError(http.StatusForbidden, errors.New("#STA-SSPI0001"))
			return
		}
		if username != "" && strings.ToLower(nego[0]) == "negotiate" {

			// user is computed from specific module
			stauser := models.StaUser{}
			stauser.User = username
			stauser.UserGroups = ""

			// TODO compute SID and groups
			c.Set("connection", stauser)
			c.Set("aud", "ad")
			//testhash := fmt.Sprintf("%x", md5.Sum([]byte(username)))
			//c.Set("sign_key", testhash)
		}
	}
}

func SspiHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		a, err := c.Get("aud")
		if err {
			if a == "internal" {
				return
			}
		}
		h.ServeHTTP(c.Writer, c.Request)
	}
}
