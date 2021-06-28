package ctrl

import "github.com/gin-gonic/gin"

func GetAccess(c *gin.Context) {
	c.JSON(200, gin.H{"message": "GetAccess required"})
}
