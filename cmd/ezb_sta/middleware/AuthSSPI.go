package middleware

import (
	"errors"
	"ezBastion/cmd/ezb_sta/models"
	"github.com/gin-gonic/gin"
	"github.com/quasoft/websspi"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os/user"
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
	// Handle only AD requests
	_, err := c.Get("aud")
	if err {
		return
	}
	// if request is jwt request, abort
	_, err = c.Get("jwt")
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
			ADu, err := GetUserAttributes(username)
			if err != nil {
				logg.Error("user error #SSPI0002: " + authHead)
				c.AbortWithError(http.StatusForbidden, errors.New("#STA-SSPI0002"))
			}
			stauser.UserSid = ADu.Uid
			// get the groups

			groups, err := ADu.GroupIds()
			if err != nil {
				logg.Error("user error when getting groups #SSPI0003: " + authHead)
				c.AbortWithError(http.StatusForbidden, errors.New("#STA-SSPI0003"))
			}
			// Parse all groupids to get the real group names
			groupsnames := make([]string, 5)
			for _, g := range groups {
				gname, aderr := user.LookupGroupId(g)
				if aderr != nil {
					logg.Warning("GroupID " + g + " is not found in Active Directory")
					continue
				}
				groupsnames = append(groupsnames, gname.Name)
			}
			stauser.UserGroups = strings.Join(groupsnames, ",")

			// TODO compute SID and groups
			c.Set("connection", stauser)
			c.Set("aud", "ad")

		}
	}
}

func SspiHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := c.Get("aud")
		if err {
			return
		}
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func GetUserAttributes(username string) (u *user.User, ret error) {
	u, err := user.Lookup(username)
	if err != nil {
		log.Fatal(err.Error())
	}

	return u, ret
}
