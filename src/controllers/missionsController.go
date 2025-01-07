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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var missionsCollection *mongo.Collection = database.OpenCollection(database.Client, "missions")

func CreateMission() gin.HandlerFunc {
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

		userType := claims["UserType"].(string)

		if userType != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário sem permissão", "tipo": userType})
			return
		}

		var mission model.Missions

		if err := c.ShouldBindJSON(&mission); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if mission.Duration < 7200000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A duração da missão deve ser maior que 2 horas"})
			return
		}

		mission.ID = primitive.NewObjectID()
		mission.OwnerId = claims["Uid"].(string)
		mission.EndDate = time.Now().Add(time.Duration(mission.Duration) * time.Millisecond)
		mission.Completed = []string{}
		mission.CreatedAt = time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := missionsCollection.InsertOne(ctx, mission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar o post"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Missão criado com sucesso", "mission": mission})
	}
}

func GetMissions() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := missionsCollection.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"createdAt": -1}))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar missões"})
			return
		}
		defer cursor.Close(ctx)

		var missions []model.Missions

		if err = cursor.All(ctx, &missions); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar missões"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"missions": missions})
	}
}

func GetCurrentMissions() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.M{
			"endDate": bson.M{"$gt": time.Now()},
		}

		cursor, err := missionsCollection.Find(ctx, filter, options.Find().SetSort(bson.M{"createdAt": -1}))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar missões"})
			return
		}
		defer cursor.Close(ctx)

		var missions []model.Missions

		if err := cursor.All(ctx, &missions); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar missões"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"missions": missions})
	}
}

func CompleteMission() gin.HandlerFunc {
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

		userType := claims["UserType"].(string)
		missionIdParam := c.Param("missionId")
		userId := c.Param("userId")

		if userType != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário sem permissão"})
			return
		}

		missionId, err := primitive.ObjectIDFromHex(missionIdParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID da missão inválido"})
			return
		}

		var mission model.Missions

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = missionsCollection.FindOne(ctx, bson.M{"_id": missionId}).Decode(&mission)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Missão não encontrada"})
			return
		}

		if mission.EndDate.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A missão já expirou"})
			return
		}

		if contains(mission.Completed, userId) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Usuário já completou essa missão"})
			return
		}

		mission.Completed = append(mission.Completed, userId)

		_, err = missionsCollection.UpdateOne(ctx, bson.M{"_id": missionId}, bson.M{"$set": bson.M{"completed": mission.Completed}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar missão"})
			return
		}

		userCollection := database.OpenCollection(database.Client, "users")
		_, err = userCollection.UpdateOne(
			ctx,
			bson.M{"uid": userId},
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

		c.JSON(http.StatusOK, gin.H{"message": "Missão completada com sucesso"})
	}
}

func DeleteMission() gin.HandlerFunc {
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

		userType := claims["UserType"].(string)

		if userType != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário sem permissão"})
			return
		}

		missionIdParam := c.Param("missionId")

		missionId, err := primitive.ObjectIDFromHex(missionIdParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID do post inválido"})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err = missionsCollection.DeleteOne(ctx, bson.M{"_id": missionId})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar missão"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Missão deletada com sucesso"})
	}
}