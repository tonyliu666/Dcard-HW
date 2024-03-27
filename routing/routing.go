package routing

import (
	"dcardapp/service"

	"github.com/gin-gonic/gin"
)

// deal with the user routes
func AddUserRouter(r *gin.RouterGroup) {
	r.POST("/ad", service.CreateADs)
	r.GET("/ad", service.GetADs)
	r.GET("/ad/:offset/:limit/:age/:gender/:country/:platform", service.GetADsWithConditions)
}
