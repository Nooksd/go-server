package routes

import (
	controller "github.com/Nooksd/go-server/src/controllers"
	middleware "github.com/Nooksd/go-server/src/middlewares"
	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {
	router.Use(middleware.Authenticate())
	router.POST("/user/create", controller.CreateUser())
	router.GET("/users", controller.SearchUsers())
	router.POST("/avatar/upload/:userId", controller.UploadAvatar())
	router.GET("/users/birthdays", controller.GetBirthdays())
	router.GET("/users/:userId", controller.GetOneUser())
	router.GET("/users/get-current-user", controller.GetCurrentUser())
	router.PUT("/users/update/:userId", controller.UpdateOneUser())
}
