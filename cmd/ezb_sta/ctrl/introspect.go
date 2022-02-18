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
			c.AbortWithError(http.StatusNoContent, errors.New("#STA-INSP0003"))
		}
		user, ok := ldapclient.JTIMap[fmt.Sprintf("%s", j)]
		if ok {
			username := strings.Split(user, "\\")
			u := new(models.IntrospectUser)
			ADobj, err := middleware.F_GetADproperties(username[1], ldapclient)
			if err != nil {
				c.AbortWithError(http.StatusConflict, errors.New("#STA-INSP0002"))
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

			c.JSON(http.StatusOK, u)
		} else {
			c.AbortWithError(http.StatusNoContent, errors.New("#STA-INSP0001"))
		}
	}

}
