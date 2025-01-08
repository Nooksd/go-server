package controllers

import (
	"context"
	"net/http"
	"time"

	database "github.com/Nooksd/go-server/src/db"
	model "github.com/Nooksd/go-server/src/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var validationCollection = database.OpenCollection(database.Client, "validations")

func CreateValidation() gin.HandlerFunc {
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

		var validation model.Validation

		if err := c.ShouldBindJSON(&validation); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
			return
		}

		validation.ID = primitive.NewObjectID()
		validation.UserID = userId
		validation.Status = "pending"
		validation.SubmittedAt = time.Now()

		validationErrors := validate.Struct(validation)
		if validationErrors != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErrors.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := validationCollection.InsertOne(ctx, validation)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar validação"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Validação criada com sucesso", "validation": validation})
	}
}

func AcceptValidation() gin.HandlerFunc {
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

		if claims["UserType"].(string) != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Permissão negada"})
			return
		}

		validationId := c.Param("validationId")

		validationID, err := primitive.ObjectIDFromHex(validationId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de validação inválido"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var validation model.Validation
		err = validationCollection.FindOne(ctx, bson.M{"_id": validationID, "status": "pending"}).Decode(&validation)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Validação não encontrada ou já processada"})
			return
		}

		missionId := validation.MissionID

		missionID, err := primitive.ObjectIDFromHex(missionId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de validação inválido"})
			return
		}

		var mission model.Missions
		err = missionsCollection.FindOne(ctx, bson.M{"_id": missionID}).Decode(&mission)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Missão associada não encontrada"})
			return
		}

		if mission.EndDate.Before(validation.SubmittedAt) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Data de entrega da missão expirou, não é possível aceitar a validação."})
			return
		}

		if contains(mission.Completed, validation.UserID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Usuário já completou essa missão"})
			return
		}

		_, err = validationCollection.UpdateOne(
			ctx,
			bson.M{"_id": validationID},
			bson.M{
				"$set": bson.M{
					"status":      "validated",
					"validatedBy": claims["Uid"].(string),
					"validatedAt": time.Now(),
				},
			},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao aceitar validação"})
			return
		}

		mission.Completed = append(mission.Completed, validation.UserID)
		_, err = missionsCollection.UpdateOne(ctx, bson.M{"_id": mission.ID}, bson.M{"$set": bson.M{"completed": mission.Completed}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar missão"})
			return
		}

		userCollection := database.OpenCollection(database.Client, "users")
		_, err = userCollection.UpdateOne(
			ctx,
			bson.M{"uid": validation.UserID},
			bson.M{
				"$inc": bson.M{
					"pTotal": +mission.Value,
				},
			},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar o PTotal do usuário"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Validação aceita e progresso atualizado com sucesso"})
	}
}

func RejectValidation() gin.HandlerFunc {
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

		if claims["UserType"].(string) != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Permissão negada"})
			return
		}

		validationID, err := primitive.ObjectIDFromHex(c.Param("validationId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de validação inválido"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := validationCollection.UpdateOne(
			ctx,
			bson.M{"_id": validationID, "status": "pending"},
			bson.M{
				"$set": bson.M{
					"status":      "rejected",
					"validatedBy": claims["Uid"].(string),
					"validatedAt": time.Now(),
				},
			},
		)
		if err != nil || result.ModifiedCount == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao rejeitar validação"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Validação rejeitada com sucesso"})
	}
}

func GetPendingValidations() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := validationCollection.Find(ctx, bson.M{"status": "pending"})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar validações pendentes"})
			return
		}
		defer cursor.Close(ctx)

		var validations []model.Validation
		if err = cursor.All(ctx, &validations); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar validações pendentes"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"validations": validations})
	}
}
