package routes

import (
	controller "github.com/Nooksd/go-server/src/controllers"
	middleware "github.com/Nooksd/go-server/src/middlewares"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controller.GetAllUsers())
	incomingRoutes.GET("/users/:userId", controller.GetOneUser())
	incomingRoutes.GET("/users/get-current-user", controller.GetCurrentUser())
}
