package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	database "github.com/Nooksd/go-server/src/db"
	helper "github.com/Nooksd/go-server/src/helpers"
	model "github.com/Nooksd/go-server/src/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "users")
var validate = validator.New()

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
		return ""
	}
	return string(hashedPassword)
}

func VerifyPassword(providedPassword string, storedHash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(providedPassword))
	if err != nil {
		return fmt.Errorf("email ou senha incorretos")
	}
	return nil
}

func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user model.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "erro ao ler dados"})
			return
		}

		validationErrors := validate.Struct(user)
		if validationErrors != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErrors.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "usuário já existe"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password
		user.ID = primitive.NewObjectID()
		user.Uid = user.ID.Hex()

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "usuário não cadastrado"})
			return
		}
		defer cancel()

		c.JSON(http.StatusCreated, resultInsertionNumber)

	}
}

func LoginUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user model.User
		var foundUser model.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "erro ao ler dados"})
			return
		}
		if (user.Email == nil) || (user.Password == nil) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email e senha são obrigatórios"})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "email e/ou senha incorretos"})
			return
		}

		if err := VerifyPassword(*user.Password, *foundUser.Password); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "email e/ou senha incorretos"})
			return
		}

		accessToken, refreshToken, _ := helper.GenerateTokens(*foundUser.Email, *foundUser.Name, foundUser.Uid, *foundUser.UserType, true)

		c.JSON(http.StatusOK, gin.H{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
			"user":         foundUser,
		})
	}
}

func GetCurrentUser() gin.HandlerFunc {
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

		userID := claims["Uid"].(string)
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.M{"uid": userID}

		var user model.User
		err := userCollection.FindOne(ctx, filter).Decode(&user)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Sucesso",
			"user":    user,
		})
	}
}

func UpdateOneUser() gin.HandlerFunc {
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
		targetUserId := c.Param("userId")

		if userType != "ADMIN" && userId != targetUserId {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Você não tem permissão para atualizar este usuário"})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var userUpdates bson.M
		if err := c.BindJSON(&userUpdates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao ler os dados de atualização"})
			return
		}

		delete(userUpdates, "userType")
		delete(userUpdates, "password")
		delete(userUpdates, "uid")

		filter := bson.M{"uid": targetUserId}
		update := bson.M{"$set": userUpdates}

		result, err := userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar os dados do usuário"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
			return
		}

		var userProfile model.User
		err = userCollection.FindOne(ctx, filter).Decode(&userProfile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar o usuário atualizado"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Usuário atualizado com sucesso",
			"result":  result,
			"user":    userProfile,
		})
	}
}

func GetAllUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var users []model.User

		cursor, err := userCollection.Find(ctx, bson.M{})
		if err != nil {
			log.Println("Erro ao buscar usuários:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar usuários"})
			return
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var user model.User
			if err := cursor.Decode(&user); err != nil {
				log.Println("Erro ao decodificar usuário:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar usuários 1"})
				return
			}
			users = append(users, user)
		}

		if err := cursor.Err(); err != nil {
			log.Println("Erro no cursor:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar usuários 2"})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

func GetOneUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("userId")

		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100+time.Second)

		var user model.User
		err := userCollection.FindOne(ctx, bson.M{"uid": userId}).Decode(&user)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "usuário não encontrado", "erro": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
