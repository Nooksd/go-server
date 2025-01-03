package routes

import (
	controller "github.com/Nooksd/go-server/src/controllers"
	middleware "github.com/Nooksd/go-server/src/middlewares"
	"github.com/gin-gonic/gin"
)

func PostRoutes(router *gin.Engine) {
	router.POST("/post/create", middleware.Authenticate(), controller.UploadPost())
	router.GET("/post/get", middleware.Authenticate(), controller.GetPosts())

	router.POST("/post/like/:postId", middleware.Authenticate(), controller.LikePost())
	router.POST("/post/dislike/:postId", middleware.Authenticate(), controller.DislikePost())
	router.POST("/post/comment/:postId", middleware.Authenticate(), controller.CommentPost())
	router.DELETE("/post/comment/delete/:postId/:commentId", middleware.Authenticate(), controller.DeleteComment())

	router.POST("/post/image/upload", middleware.Authenticate(), controller.UploadImage())
	router.GET("/post/image/get/:image", controller.GetImage())

	router.DELETE("/post/delete/:postId", middleware.Authenticate(), controller.DeletePost())
}
