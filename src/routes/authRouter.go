package routes

import (
	controller "github.com/Nooksd/go-server/src/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/users/create", controller.CreateUser())
	incomingRoutes.POST("/users/login", controller.LoginUser())
}
