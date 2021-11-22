package middleware

import (
	"encoding/base64"
	"errors"
	"ezBastion/cmd/ezb_sta/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jtblin/go-ldap-client"
	log "github.com/sirupsen/logrus"
	genldap "gopkg.in/ldap.v2"
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

			ok, attr, err := ldapclient.Authenticate(username, password)
			if ok {
				// user is computed from specific module
				stauser := models.StaUser{}
				// Compute the group list
				groupsnames, aderr := ldapclient.GetGroupsOfUser(attr["distinguishedName"])
				if aderr != nil {
					log.Errorf("Error when getting groups for user %s ", username)
				} else {
					stauser.UserGroups = groupsnames
				}
				stauser.User = username
				// TODO compute SID and groups
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

func F_GetGroupsOfUser(userdn string, lc *ldap.LDAPClient) ([]string, error) {

	searchRequest := genldap.NewSearchRequest(
		lc.Base,
		genldap.ScopeWholeSubtree, genldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(lc.GroupFilter, userdn),
		[]string{"cn"}, // can it be something else than "cn"?
		nil,
	)
	sr, err := lc.Conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	groups := []string{}
	for _, entry := range sr.Entries {
		groups = append(groups, entry.GetAttributeValue("cn"))
	}

	return groups, nil
}
