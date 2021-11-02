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

		CheckUserandSet(c, username, password, nil)
	}
}
