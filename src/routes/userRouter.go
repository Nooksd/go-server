package routes

import (
	controller "github.com/Nooksd/go-server/src/controllers"
	// middleware "github.com/Nooksd/go-server/src/middlewares"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	// incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.POST("/users", controller.GetAllUsers())
	incomingRoutes.POST("/users/:userId", controller.GetOneUser())
}
