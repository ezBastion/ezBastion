package ctrl

import (
	"errors"
	"ezBastion/cmd/ezb_sta/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Introspect() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.ParseForm()
		kform := "jti"
		vform := c.Request.Form[kform]
		if len(vform) != 1 {
			c.AbortWithStatusJSON(http.StatusBadRequest, errors.New("#STA0200 - wrong jti request"))
		}
		// TODO how to compute the introspect part for a next feature
		// Compute a DUMMY instrospect
		u := new(models.IntrospectUser)
		u.Samaccountname = "Dummy"
		u.Groups = []string{"Dummy group A", "Dummy group B"}
		c.JSON(http.StatusOK, u)

	}

}
