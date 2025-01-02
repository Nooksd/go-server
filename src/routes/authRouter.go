package routes

import (
	controller "github.com/Nooksd/go-server/src/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	router.POST("/auth/create", controller.CreateUser())
	router.POST("/auth/login", controller.LoginUser())
	router.GET("/auth/refresh-token", controller.RefreshToken())
}
