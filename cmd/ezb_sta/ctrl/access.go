package ctrl

import "github.com/gin-gonic/gin"

func Access() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Access required"})
	}
}
