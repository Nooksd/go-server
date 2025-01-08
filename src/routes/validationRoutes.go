package routes

import (
	controller "github.com/Nooksd/go-server/src/controllers"
	"github.com/gin-gonic/gin"
)

func ValidationRoutes(router *gin.Engine) {
	router.POST("/validation/create", controller.CreateValidation())
	router.PUT("/validation/accept/:validationId", controller.AcceptValidation())
	router.PUT("/validation/reject/:validationId", controller.RejectValidation())
	router.GET("/validation/get-pending", controller.GetPendingValidations())
}
