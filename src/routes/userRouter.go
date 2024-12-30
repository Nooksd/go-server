package routes

import (
	controller "github.com/Nooksd/go-server/src/controllers"
	middleware "github.com/Nooksd/go-server/src/middlewares"
	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {
	router.Use(middleware.Authenticate())
	router.GET("/users", controller.GetAllUsers())
	router.GET("/users/:userId", controller.GetOneUser())
	router.GET("/users/get-current-user", controller.GetCurrentUser())
	router.PUT("/users/update/:userId", controller.UpdateOneUser())
}
