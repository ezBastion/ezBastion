package middleware

import (
	"ezBastion/cmd/ezb_sta/models"
	"github.com/gin-gonic/gin"
)

func EzbAuthForm(c *gin.Context) {

	var mp models.EzbFormAuth
	err := c.ShouldBindJSON(&mp)
	if err != nil {
		// Error in binding, this is not an basic
		return
	}

	if mp.Grant_type == "password" {

		username := mp.Username
		password := mp.Password
		err := checkDBUser(c, username, password)
		if err == 0 {
			// user is computed from specific module
			stauser := models.StaUser{}
			stauser.User = username
			c.Set("connection", stauser)
			c.Set("aud", "internal")
		}
	}
}
