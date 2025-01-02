package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

var postCollection *mongo.Collection = database.OpenCollection(database.Client, "posts")

func UploadPost() gin.HandlerFunc {
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

		var post model.Post

		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if len(post.Hashtags) > 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "O número máximo de hashtags permitido é 3"})
			return
		}

		post.ID = primitive.NewObjectID()
		post.Name = claims["Name"].(string)
		post.OwnerId = claims["Uid"].(string)
		post.AvatarURL = claims["ProfilePictureUrl"].(string)
		post.Likes = 0
		post.Comments = []model.Comment{}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := postCollection.InsertOne(ctx, post)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar o post"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Post criado com sucesso", "post": post})
	}
}

func GetPosts() gin.HandlerFunc {
	return func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		pageSize := 5

		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Número da página inválido"})
			return
		}

		skip := (pageInt - 1) * pageSize

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := postCollection.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"_id": -1}).SetSkip(int64(skip)).SetLimit(int64(pageSize)))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar posts"})
			return
		}
		defer cursor.Close(ctx)

		var posts []model.Post
		if err = cursor.All(ctx, &posts); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar posts"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"posts": posts, "page": pageInt})
	}
}

func DeletePost() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func UploadImage() gin.HandlerFunc {
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

		err := c.Request.ParseMultipartForm(10 << 20)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar o arquivo"})
			return
		}

		file, _, err := c.Request.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhum arquivo enviado"})
			return
		}
		defer file.Close()

		timeStamp := time.Now().Unix()
		filename := fmt.Sprintf("%s_%d.jpg", userId, timeStamp)

		filePath := filepath.Join("..", "uploads", "post", filename)

		err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar diretório"})
			return
		}

		dst, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar o arquivo"})
			return
		}
		defer dst.Close()

		_, err = dst.ReadFrom(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao escrever o arquivo"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"url": "http://localhost:9000/post/image/get/" + filename})
	}
}

func GetImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		image := c.Param("image")

		filePath := filepath.Join("..", "uploads", "post", image)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Imagem não encontrada"})
			return
		}

		c.File(filePath)
	}
}
