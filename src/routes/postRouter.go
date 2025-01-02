package routes

import (
	controller "github.com/Nooksd/go-server/src/controllers"
	"github.com/gin-gonic/gin"
)

func PostRoutes(router *gin.Engine) {
	router.POST("/post/upload", controller.UploadPost())
	router.GET("/post/get", controller.GetPosts())

	router.POST("/post/image/upload", controller.UploadImage())
	router.GET("/post/image/get/:image", controller.GetImage())

	router.DELETE("/post/delete/:postId", controller.DeletePost())
}
