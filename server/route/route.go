package route

import (
	"github.com/gin-gonic/gin"
	"summerflower.local/picture_api/server/service"
)

func SetupRouter(router *gin.Engine) {
	images := router.Group("/images")
	{
		images.GET("/", service.GetRandom)
		images.GET("/:id", service.GetImageById)
	}
	router.StaticFile("/favicon.ico", "resources/public/favicon.ico")
}
