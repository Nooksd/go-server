package routes

import (
	controller "github.com/Nooksd/go-server/src/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	router.POST("/users/create", controller.CreateUser())
	router.POST("/users/login", controller.LoginUser())
}
