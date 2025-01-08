package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
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

		validationErrors := validate.Struct(mission)
		if validationErrors != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErrors.Error()})
			return
		}

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

		date := time.Now().UTC()

		filter := bson.M{
			"endDate": bson.M{"$gt": date},
		}

		cursor, err := missionsCollection.Find(ctx, filter, options.Find().SetSort(bson.M{"endDate": 1}))
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

func VerifyCompletion() gin.HandlerFunc {
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
		missionIdParam := c.Param("missionId")

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

		var completed bool = false
		switch mission.MissionType {
		case "FEEDPOST":
			completed = verifyFeedPost(mission, userId)
		case "FEEDHASHTAG":
			completed = verifyFeedHashtag(mission, userId)
		case "FEEDIMAGE":
			completed = verifyFeedImage(userId)
		case "INSTAGRAMSTORY":
			completed, err = verifyInstagramMention(userId)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de missão inválido"})
			return
		}

		if !completed {
			c.JSON(http.StatusPreconditionFailed, gin.H{"error": "Missão não concluida", "err": err})
			return
		}

		mission.Completed = append(mission.Completed, userId)

		_, err = missionsCollection.UpdateOne(ctx, bson.M{"_id": mission.ID}, bson.M{"$set": bson.M{"completed": mission.Completed}})
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

func verifyFeedPost(mission model.Missions, userId string) bool {
	postCollection := database.OpenCollection(database.Client, "posts")

	filter := bson.M{"ownerId": userId}
	var lastPost model.Post

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := postCollection.FindOne(ctx, filter, options.FindOne().SetSort(bson.M{"createdAt": -1})).Decode(&lastPost)
	if err != nil {
		return false
	}

	return lastPost.CreatedAt.After(mission.CreatedAt)
}

func verifyFeedHashtag(mission model.Missions, userId string) bool {
	postCollection := database.OpenCollection(database.Client, "posts")

	var lastPost model.Post
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.FindOne()
	findOptions.Sort = bson.M{"createdAt": -1}

	err := postCollection.FindOne(ctx, bson.M{"ownerId": userId}, findOptions).Decode(&lastPost)
	if err != nil {
		return false
	}

	if lastPost.CreatedAt.Before(mission.CreatedAt) {
		return false
	}

	if mission.Hashtag == "" {
		return true
	}

	for _, postHashtag := range lastPost.Hashtags {
		if postHashtag == mission.Hashtag {
			return true
		}
	}

	return false
}

func verifyFeedImage(userId string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	postCollection := database.OpenCollection(database.Client, "posts")
	findOptions := options.FindOne()
	findOptions.Sort = bson.M{"createdAt": -1}

	var lastPost model.Post
	err := postCollection.FindOne(ctx, bson.M{"ownerId": userId}, findOptions).Decode(&lastPost)
	if err != nil {
		return false
	}

	return lastPost.ImageUrl != ""
}

func verifyInstagramMention(userId string) (bool, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user model.User
	userCollection := database.OpenCollection(database.Client, "users")

	err := userCollection.FindOne(ctx, bson.M{"uid": userId}).Decode(&user)
	if err != nil {
		return false, fmt.Errorf("erro ao buscar usuário: %v", err)
	}

	if user.InstagramURL == nil || *user.InstagramURL == "" {
		return false, fmt.Errorf("usuário não possui Instagram vinculado")
	}

	accessToken := os.Getenv("INSTAGRAM_ACCESS_TOKEN")
	instaProfileID := os.Getenv("INSTAGRAM_ID")

	posts, err := getInstagramPosts(instaProfileID, accessToken)
	if err != nil {
		return false, fmt.Errorf("erro ao buscar posts do Instagram: %v", err)
	}

	for _, post := range posts {
		if strings.Contains(post.Caption, "@sd_nook") {
			return true, nil
		}
	}

	return false, nil
}

func getInstagramPosts(instagramUserID, accessToken string) ([]struct {
	Caption string `json:"caption"`
}, error) {
	resp, err := http.Get(fmt.Sprintf("https://graph.instagram.com/v16.0/%s/media?fields=caption&access_token=%s", instagramUserID, accessToken))
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Caption string `json:"caption"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result.Data, nil
}
