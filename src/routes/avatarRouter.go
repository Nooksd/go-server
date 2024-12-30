package routes

import (
	controller "github.com/Nooksd/go-server/src/controllers"
	middleware "github.com/Nooksd/go-server/src/middlewares"
	"github.com/gin-gonic/gin"
)

func AvatarRoutes(router *gin.Engine) {
	router.POST("/avatar/upload/:userId", middleware.Authenticate(), controller.UploadAvatar())
	router.GET("/avatar/get/:userId", controller.GetAvatar())
}
