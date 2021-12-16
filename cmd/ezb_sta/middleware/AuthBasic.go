package middleware

import (
	"encoding/base64"
	"errors"
	"ezBastion/cmd/ezb_sta/models"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func EzbAuthBasic(ldapclient *models.Ldapinfo) gin.HandlerFunc {
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

			ok, _, err := LDAPauth(ldapclient, username, password)
			if ok {
				// user is computed from specific module
				stauser := models.StaUser{}
				// Compute the group list
				/*groupsnames, aderr := ldapclient.GetGroupsOfUser(attr["distinguishedName"])
				if aderr != nil {
					log.Errorf("Error when getting groups for user %s ", username)
				} else {
					stauser.UserGroups = groupsnames
				}*/
				stauser.User = username
				attr, err := F_GetADproperties(username, ldapclient)
				if err != nil {
					logg.Warning(fmt.Sprintf("Active Directory properties retrieved errors for user %s", username))
				} else {
					stauser.ExtProperties = attr
					//stauser.ExtProperties.Groups = groupsnames
				}
				c.Set("connection", stauser)
				c.Set("aud", "ad")
			} else {
				if err != nil {
					log.Errorf("Error authenticating user %s: %+v", "username", err)
				}
			}
		}
	}
}
