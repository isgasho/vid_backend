package group

import (
	. "vid/app/controller"
	"vid/app/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRawGroup(router *gin.Engine) {

	jwt := middleware.JWTMiddleware(false)

	rawGroup := router.Group("/raw")
	{
		rawGroup.GET("/image/:user/:filename", RawCtrl.RawImage)
		rawGroup.GET("/video/:user/:filename", RawCtrl.RawVideo)
		uploadSubGroup := rawGroup.Group("/upload")
		{
			uploadSubGroup.Use(jwt)
			uploadSubGroup.POST("/image", RawCtrl.UploadImage)
			uploadSubGroup.POST("/video", RawCtrl.UploadVideo)
		}
	}
}