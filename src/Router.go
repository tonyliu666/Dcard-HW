package src

import (
	"github.com/gin-gonic/gin"
)

func AddUserRouter(r *gin.RouterGroup) {
	user := r.Group("/ad")

	user.GET("/user", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "user",
		})
	})
}

