package controllers

import (
	"context"
	"net/http"
	"time"

	database "github.com/Nooksd/go-server/src/db"
	models "github.com/Nooksd/go-server/src/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var tokenCollection = database.OpenCollection(database.Client, "deviceTokens")

func SaveDeviceToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		userClaims, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
			return
		}

		claims, ok := userClaims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar token"})
			return
		}

		var deviceToken models.DeviceToken

		if err := c.ShouldBindJSON(&deviceToken); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
			return
		}

		deviceToken.UserId = claims["Uid"].(string)

		if len(deviceToken.NotificationTypes) == 0 {
			deviceToken.NotificationTypes = []string{"feed", "birthday", "contact", "mission", "all"}
		}

		validationErrors := validate.Struct(deviceToken)
		if validationErrors != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErrors.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.M{"userId": deviceToken.UserId, "deviceToken": deviceToken.DeviceToken}

		update := bson.M{
			"$set": bson.M{
				"userId":            deviceToken.UserId,
				"deviceToken":       deviceToken.DeviceToken,
				"notificationTypes": deviceToken.NotificationTypes,
			},
		}
		opts := options.Update().SetUpsert(true)

		_, err := tokenCollection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar token do dispositivo"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Token salvo com sucesso"})
	}
}
