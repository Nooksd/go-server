package helpers

import (
	"log"
	"os"
	"time"

	database "github.com/Nooksd/go-server/src/db"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

type SignedDetails struct {
	Email    string
	Name     string
	Uid      string
	UserType string
	jwt.RegisteredClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateTokens(email string, name string, uid string, userType string, keepLogged bool) (signedAccessToken string, signedRefreshToken string, err error) {
	accessTokenDuration := time.Hour * 24
	refreshTokenDuration := time.Hour * 24 * 7

	accessClaims := &SignedDetails{
		Email:    email,
		Name:     name,
		Uid:      uid,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshClaims := &SignedDetails{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Println("Erro ao criar Access Token:", err)
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Println("Erro ao criar Refresh Token:", err)
		return "", "", err
	}

	if keepLogged {
		return accessToken, refreshToken, nil
	}

	return accessToken, "", nil
}