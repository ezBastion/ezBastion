package ctrl

import (
	"errors"
	"ezBastion/cmd/ezb_sta/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func Memberof() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.ParseForm()
		kform := "group"
		vform := c.Request.Form[kform]
		if len(vform) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, errors.New("#STA0100 - memberof requested without any group name"))
		}
		stauser := c.MustGet("connection").(models.StaUser)
		for _, gname := range vform {
			if len(stauser.UserGroups) != 0 {
				for _, groups := range stauser.UserGroups {
					if strings.ToLower(groups) == strings.ToLower(gname) {
						c.JSON(http.StatusOK, "1")
						return
					}
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, errors.New("#STA0199 - user is not member of the requested group"))
	}
}
