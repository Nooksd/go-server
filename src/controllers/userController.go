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

		defaultRole := "Membro"
		emptyString := ""
		defaultPoints := 0

		user.ProfilePictureUrl = "http://192.168.1.68:9000/avatar/get/avatar1"
		user.PhoneNumber = &emptyString
		user.Role = &defaultRole
		user.EntryDate = time.Now()
		user.Birthday = time.Now()
		user.LinkedinURL = &emptyString
		user.FacebookURL = &emptyString
		user.InstagramURL = &emptyString
		user.PTotal = &defaultPoints
		user.PSpent = &defaultPoints
		user.PCurrent = &defaultPoints

		user.ID = primitive.NewObjectID()
		user.Uid = user.ID.Hex()

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "usuário não cadastrado"})
			return
		}
		defer cancel()

		helper.CreateNotification(
			fmt.Sprintf(
				"Novo funcionário adicionado: %s",
				*user.Name,
			),
			"contact")

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

		accessToken, refreshToken, _ := helper.GenerateTokens(*foundUser.Email, *foundUser.Name, foundUser.ProfilePictureUrl, *foundUser.Role, foundUser.Uid, *foundUser.UserType, true)

		c.JSON(http.StatusOK, gin.H{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
			"user":         foundUser,
			"type":         foundUser.UserType,
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

func RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		refreshToken := c.GetHeader("Token")
		if refreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token não fornecido"})
			return
		}

		token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(helper.SECRET_KEY), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token inválido"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar token"})
			return
		}

		email := claims["Email"].(string)
		name := claims["Name"].(string)
		profilePictureUrl := claims["ProfilePictureUrl"].(string)
		role := claims["Role"].(string)
		uid := claims["Uid"].(string)
		userType := claims["UserType"].(string)

		newAccessToken, _, err := helper.GenerateTokens(email, name, profilePictureUrl, role, uid, userType, false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar novo token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"accessToken": newAccessToken,
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

func SearchUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.DefaultQuery("name", "")

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var users []model.User

		cursor, err := userCollection.Find(ctx, bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: ".*" + name + ".*", Options: "i"}}})
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
			user.Password = nil
			user.Email = nil
			users = append(users, user)
		}

		if err := cursor.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar usuários"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"users": users})
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

		user.Password = nil
		user.Email = nil

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "usuário não encontrado", "erro": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)
	}

}

func GetBirthdays() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		currentDate := time.Now()
		currentMonthDay := fmt.Sprintf("%02d-%02d", currentDate.Month(), currentDate.Day())

		pipeline := mongo.Pipeline{
			{{Key: "$addFields", Value: bson.M{
				"birthdaySortable": bson.M{
					"$substr": []interface{}{"$birthday", 5, 5},
				},
			}}},
			{{Key: "$match", Value: bson.M{
				"birthdaySortable": bson.M{"$gte": currentMonthDay},
			}}},
			{{Key: "$sort", Value: bson.M{"birthdaySortable": 1}}},
			{{Key: "$project", Value: bson.M{
				"name":              1,
				"profilePictureUrl": 1,
				"role":              1,
				"birthday":          1,
				"_id":               0,
			}}},
		}

		cursor, err := userCollection.Aggregate(ctx, pipeline)
		if err != nil {
			log.Printf("Erro ao executar agregação: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar aniversários"})
			return
		}
		defer cursor.Close(ctx)

		var users []bson.M
		if err := cursor.All(ctx, &users); err != nil {
			log.Printf("Erro ao decodificar usuários: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar dados dos usuários"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"birthdays": users})
	}
}
