package routes

import (
	"github.com/Nooksd/go-server/src/controllers"
	"github.com/gin-gonic/gin"
)

func NotificationRoutes(router *gin.Engine) {
	router.POST("/notification/create", controllers.CreateNotification())
	router.GET("/notification/get-all", controllers.GetNotifications())
	router.PUT("/notification/read/:notificationId", controllers.ReadNotification())
	router.DELETE("/notification/delete/:notificationId", controllers.DeleteNotification())
	router.POST("/notification/register-token", controllers.SaveDeviceToken())

}
