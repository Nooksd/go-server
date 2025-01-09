package controllers

import (
	"context"
	"net/http"
	"time"

	database "github.com/Nooksd/go-server/src/db"
	"github.com/Nooksd/go-server/src/helpers"
	models "github.com/Nooksd/go-server/src/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var notificationCollection = database.OpenCollection(database.Client, "notifications")

func CreateNotification() gin.HandlerFunc {
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

		var userType = claims["UserType"].(string)

		if userType != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário sem permissão"})
			return
		}

		var notificationRequest struct {
			Text string `json:"text" validate:"required"`
			Type string `json:"type" validate:"required"`
		}

		if err := c.ShouldBindJSON(&notificationRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
			return
		}

		err := helpers.CreateNotification(notificationRequest.Text, notificationRequest.Type)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar e enviar notificação"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Notificação criada e push enviado com sucesso"})
	}
}

func GetNotifications() gin.HandlerFunc {
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

		userId := claims["Uid"].(string)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.M{
			"$or": []bson.M{
				{"type": bson.M{"$not": bson.M{"$regex": "^.{24}$"}}},
				{"type": userId},
			},
		}

		cursor, err := notificationCollection.Find(ctx, filter, options.Find().SetSort(bson.M{"createdAt": -1}).SetLimit(50))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar notificações"})
			return
		}

		var notifications []models.Notification

		if err = cursor.All(ctx, &notifications); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar notificações"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"notifications": notifications})
	}
}

func ReadNotification() gin.HandlerFunc {
	return func(c *gin.Context) {
		notificationId := c.Param("notificationId")

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

		userId := claims["Uid"].(string)

		notificationObjId, err := primitive.ObjectIDFromHex(notificationId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID da notificação inválido"})
			return
		}

		var notification models.Notification

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = notificationCollection.FindOne(ctx, bson.M{"_id": notificationObjId}).Decode(&notification)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notificação nao encontrada"})
			return
		}

		if contains(notification.Visualized, userId) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Notificação já marcada como lida"})
			return
		}

		notification.Visualized = append(notification.Visualized, userId)

		_, err = notificationCollection.UpdateOne(ctx, bson.M{"_id": notificationObjId}, bson.M{"$set": bson.M{"visualized": notification.Visualized}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar notificação"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Notificação marcada como lida"})
	}
}

func DeleteNotification() gin.HandlerFunc {
	return func(c *gin.Context) {
		notificationId := c.Param("notificationId")

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

		var userType = claims["UserType"].(string)

		if userType != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário sem permissão"})
			return
		}

		notificationObjId, err := primitive.ObjectIDFromHex(notificationId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID da notificação inválido"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := notificationCollection.DeleteOne(ctx, bson.M{"_id": notificationObjId})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar notificação"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notificação não encontrada"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Notificação deletada com sucesso"})
	}
}
