package routes

import (
	controller "github.com/Nooksd/go-server/src/controllers"
	"github.com/gin-gonic/gin"
)

func PostRoutes(router *gin.Engine) {
	router.POST("/post/create", controller.UploadPost())
	router.GET("/post/get/:postId", controller.GetPost())
	router.GET("/post/get", controller.GetPosts())

	router.POST("/post/like/:postId", controller.LikePost())
	router.POST("/post/dislike/:postId", controller.DislikePost())
	router.POST("/post/comment/:postId", controller.CommentPost())
	router.DELETE("/post/comment/delete/:postId/:commentId", controller.DeleteComment())

	router.POST("/post/image/upload", controller.UploadImage())

	router.DELETE("/post/delete/:postId", controller.DeletePost())
}
