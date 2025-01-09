package helpers

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func InitFirebaseApp() (*firebase.App, error) {
	firebaseConfig := os.Getenv("FIREBASECONFIG")
	if firebaseConfig == "" {
		log.Fatal("FIREBASECONFIG não definido no .env")
		return nil, fmt.Errorf("FIREBASECONFIG nao definido no .env")
	}
	opt := option.WithCredentialsJSON([]byte(firebaseConfig))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Erro ao inicializar Firebase App: %v", err)
		return nil, err
	}
	return app, nil
}
