package routes

import (
	controller "github.com/Nooksd/go-server/src/controllers"
	"github.com/gin-gonic/gin"
)

func ImageRoutes(router *gin.Engine) {
	router.GET("/avatar/get/:userId", controller.GetAvatar())
	router.GET("/post/image/get/:image", controller.GetImage())
}
