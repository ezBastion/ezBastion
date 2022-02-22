package ctrl

import (
	"errors"
	"ezBastion/cmd/ezb_sta/middleware"
	"ezBastion/cmd/ezb_sta/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func Introspect(ldapclient *models.Ldapinfo) gin.HandlerFunc {
	return func(c *gin.Context) {
		j, err := c.Get("jti")
		if err == false {
			c.AbortWithError(http.StatusNoContent, errors.New("#STA-INSP0003, no jti sent"))
		}
		user, ok := ldapclient.JTIMap[fmt.Sprintf("%s", j)]
		if ok {
			u, err := getUserFromAD(fmt.Sprintf("%s", user), ldapclient)
			if err != nil {
				c.AbortWithError(http.StatusNoContent, errors.New("#STA-INSP0002, no user found in AD"))
			}
			c.JSON(http.StatusOK, u)
		} else {
			// the jti is not in the cache map, lets compute it
			userobj, ok := c.Get("user")
			if ok == false {
				c.AbortWithError(http.StatusNoContent, errors.New("#STA-INSP0003, no user sent"))
			}
			username := fmt.Sprintf("%s", userobj)
			u, err := getUserFromAD(username, ldapclient)
			if err != nil {
				c.AbortWithError(http.StatusNoContent, errors.New("#STA-INSP0002, no user found in AD"))
			}
			c.JSON(http.StatusOK, u)

		}
	}
}

func getUserFromAD(user string, ldapclient *models.Ldapinfo) (u *models.IntrospectUser, err error) {
	username := strings.Split(user, "\\")
	u = new(models.IntrospectUser)
	ADobj, err := middleware.F_GetADproperties(username[1], ldapclient)
	if err != nil {
		return nil, errors.New("#STA-INSP0002")
	}
	u.Groups = ADobj.Groups
	u.Ou = ADobj.Ou
	u.Samaccountname = ADobj.Samaccountname
	u.Ntaccount = ADobj.Ntaccount
	u.Givenname = ADobj.Givenname
	u.Emailaddress = ADobj.Emailaddress
	u.Description = ADobj.Description
	u.Distinguishedname = ADobj.Distinguishedname
	u.Displayname = ADobj.Displayname

	return u, nil
}
