package ctrl

import (
	"errors"
	"ezBastion/cmd/ezb_sta/middleware"
	"ezBastion/cmd/ezb_sta/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Memberof(ldapclient *models.Ldapinfo) gin.HandlerFunc {
	return func(c *gin.Context) {
		j, err := c.Get("jti")
		if err == false {
			c.AbortWithError(http.StatusNoContent, errors.New("#STA-INSP0003, no jti sent"))
		}
		c.Request.ParseForm()
		kform := "group"
		vform := c.Request.Form[kform]
		if len(vform) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, errors.New("#STA0100 - memberof requested without any group name"))
		}
		stauser := c.MustGet("connection").(models.StaUser)
		for _, gname := range vform {
			gnameDN, err := middleware.F_GetADObjectDN(gname, ldapclient)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusConflict, fmt.Errorf("#STA0197 - Error during the DN request for the group %s ", gname))
			}
			found, err := middleware.F_GetGroupNestedMemberOf(gnameDN, stauser.User, ldapclient)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusConflict, fmt.Errorf("#STA0198 - Error, computing the nested group %s", gname))
			}
			if found {
				c.JSON(http.StatusOK, "1")
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, errors.New("#STA0199 - user is not member of the requested group"))
	}
}
