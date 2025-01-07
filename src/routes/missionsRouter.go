package routes

import (
	controller "github.com/Nooksd/go-server/src/controllers"
	"github.com/gin-gonic/gin"
)

func MissionsRoutes(router *gin.Engine) {
	router.POST("/mission/create", controller.CreateMission())
	router.GET("/mission/get-all", controller.GetMissions())
	router.GET("/mission/get-current", controller.GetCurrentMissions())
	router.PUT("/mission/complete/:missionId/:userId", controller.CompleteMission())
	router.DELETE("/mission/delete/:missionId", controller.DeleteMission())
}
