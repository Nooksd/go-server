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
	helper "github.com/Nooksd/go-server/src/helpers"
	model "github.com/Nooksd/go-server/src/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var postCollection *mongo.Collection = database.OpenCollection(database.Client, "posts")

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

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
		post.Role = claims["Role"].(string)
		post.OwnerId = claims["Uid"].(string)
		post.AvatarURL = claims["ProfilePictureUrl"].(string)
		post.Likes = []string{}
		post.Comments = []model.Comment{}
		post.CreatedAt = time.Now()

		validationErrors := validate.Struct(post)
		if validationErrors != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErrors.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := postCollection.InsertOne(ctx, post)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar o post"})
			return
		}

		helper.CreateNotification(
			fmt.Sprintf(
				"%s criou um novo post",
				post.Name,
			),
			"feed")

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

		cursor, err := postCollection.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"createdAt": -1}).SetSkip(int64(skip)).SetLimit(int64(pageSize)))
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

func GetPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		postIdParam := c.Param("postId")

		postId, err := primitive.ObjectIDFromHex(postIdParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID do post inválido"})
			return
		}

		var post model.Post

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = postCollection.FindOne(ctx, bson.M{"_id": postId}).Decode(&post)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post não encontrado"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"post": post})
	}
}

func DeletePost() gin.HandlerFunc {
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
		userId := claims["Uid"].(string)

		postIdParam := c.Param("postId")

		postId, err := primitive.ObjectIDFromHex(postIdParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID do post inválido"})
			return
		}

		var post model.Post

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = postCollection.FindOne(ctx, bson.M{"_id": postId}).Decode(&post)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post não encontrado"})
			return
		}

		if post.OwnerId != userId && userType != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Você não tem permissão para deletar este post"})
			return
		}

		_, err = postCollection.DeleteOne(ctx, bson.M{"_id": postId})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar post"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Post deletado com sucesso"})
	}
}

func LikePost() gin.HandlerFunc {
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
		postIdParam := c.Param("postId")

		postId, err := primitive.ObjectIDFromHex(postIdParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID do post inválido"})
			return
		}

		var post model.Post

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = postCollection.FindOne(ctx, bson.M{"_id": postId}).Decode(&post)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post não encontrado"})
			return
		}

		if contains(post.Likes, userId) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Usuário ja gostou do post"})
			return
		}

		post.Likes = append(post.Likes, userId)

		_, err = postCollection.UpdateOne(ctx, bson.M{"_id": postId}, bson.M{"$set": bson.M{"likes": post.Likes}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar post"})
			return
		}

		if post.OwnerId != userId {
			helper.CreateNotification(
				fmt.Sprintf(
					"%s curtiu seu post",
					claims["Name"].(string),
				),
				post.OwnerId)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Post gostado com sucesso"})
	}
}

func DislikePost() gin.HandlerFunc {
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
		postIdParam := c.Param("postId")

		postId, err := primitive.ObjectIDFromHex(postIdParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID do post inválido"})
			return
		}

		var post model.Post

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = postCollection.FindOne(ctx, bson.M{"_id": postId}).Decode(&post)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post não encontrado"})
			return
		}

		if !contains(post.Likes, userId) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Usuário ainda nao gostou do post"})
			return
		}

		var likes []string

		for _, like := range post.Likes {
			if like != userId {
				likes = append(likes, like)
			}
		}

		post.Likes = likes

		_, err = postCollection.UpdateOne(ctx, bson.M{"_id": postId}, bson.M{"$set": bson.M{"likes": post.Likes}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar post"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Post desgostado com sucesso"})
	}
}

func CommentPost() gin.HandlerFunc {
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

		postIdParam := c.Param("postId")

		postId, err := primitive.ObjectIDFromHex(postIdParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID do post inválido"})
			return
		}

		var post model.Post
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = postCollection.FindOne(ctx, bson.M{"_id": postId}).Decode(&post)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post não encontrado"})
			return
		}

		var newComment model.Comment

		if err := c.ShouldBindJSON(&newComment); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dados de comentário inválidos"})
			return
		}

		newComment.OwnerId = claims["Uid"].(string)
		newComment.Name = claims["Name"].(string)
		newComment.AvatarURL = claims["ProfilePictureUrl"].(string)
		newComment.ID = primitive.NewObjectID()

		post.Comments = append(post.Comments, newComment)

		_, err = postCollection.UpdateOne(ctx, bson.M{"_id": postId}, bson.M{"$set": bson.M{"comments": post.Comments}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao adicionar comentário"})
			return
		}

		if post.OwnerId != claims["Uid"].(string) {
			helper.CreateNotification(
				fmt.Sprintf(
					"%s comentou no seu post",
					claims["Name"].(string),
				),
				post.OwnerId)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Comentário adicionado com sucesso", "comments": post.Comments})
	}
}

func DeleteComment() gin.HandlerFunc {
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

		postIdParam := c.Param("postId")
		commentIdParam := c.Param("commentId")

		postId, err := primitive.ObjectIDFromHex(postIdParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID do post inválido"})
			return
		}

		commentId, err := primitive.ObjectIDFromHex(commentIdParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID do comentário inválido"})
			return
		}

		var post model.Post
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = postCollection.FindOne(ctx, bson.M{"_id": postId}).Decode(&post)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post não encontrado"})
			return
		}

		var commentIndex = -1
		for i, comment := range post.Comments {
			if comment.ID == commentId {
				commentIndex = i
				break
			}
		}

		if commentIndex == -1 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comentário não encontrado"})
			return
		}

		commentOwnerId := post.Comments[commentIndex].OwnerId
		userId := claims["Uid"].(string)

		if commentOwnerId != userId {
			c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para deletar este comentário"})
			return
		}

		post.Comments = append(post.Comments[:commentIndex], post.Comments[commentIndex+1:]...)

		_, err = postCollection.UpdateOne(ctx, bson.M{"_id": postId}, bson.M{"$set": bson.M{"comments": post.Comments}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar comentário"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Comentário deletado com sucesso"})
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

		filePath := filepath.Join("uploads", "post", filename)

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

		c.JSON(http.StatusOK, gin.H{"url": "http://192.168.1.68:9000/post/image/get/" + filename})
	}
}

func GetImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		image := c.Param("image")

		filePath := filepath.Join("uploads", "post", image)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Imagem não encontrada"})
			return
		}

		c.File(filePath)
	}
}
