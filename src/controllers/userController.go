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
		user.UserId = user.ID.Hex()

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

		accessToken, refreshToken, _ := helper.GenerateTokens(*foundUser.Email, *foundUser.Name, foundUser.UserId, *foundUser.UserType, true)

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

		c.JSON(http.StatusOK, gin.H{
			"message": "Sucesso",
			"user":    claims,
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar usuários"})
				return
			}
			users = append(users, user)
		}

		if err := cursor.Err(); err != nil {
			log.Println("Erro no cursor:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar usuários"})
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
		err := userCollection.FindOne(ctx, bson.M{"userid": userId}).Decode(&user)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "usuário não encontrado"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
