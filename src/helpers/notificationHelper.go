package helpers

import (
	"context"
	"log"
	"time"

	"firebase.google.com/go/v4/messaging"
	database "github.com/Nooksd/go-server/src/db"
	models "github.com/Nooksd/go-server/src/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var notificationCollection = database.OpenCollection(database.Client, "notifications")
var tokenCollection = database.OpenCollection(database.Client, "deviceTokens")

func CreateNotification(text string, notificationType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var notification models.Notification

	notification.ID = primitive.NewObjectID()
	notification.CreatedAt = time.Now()
	notification.Text = text
	notification.Type = notificationType
	notification.Visualized = []string{}

	_, err := notificationCollection.InsertOne(ctx, notification)
	if err != nil {
		log.Printf("Erro ao criar notificação: %v\n", err)
		return err
	}

	if isUID(notificationType) {
		err = SendPrivateNotification(notificationType, "Nova Notificação", text)
		if err != nil {
			log.Printf("Erro ao enviar notificação privada: %v\n", err)
			return err
		}
		log.Println("Notificação privada enviada com sucesso")
		return nil
	}

	filter := bson.M{"notificationTypes": bson.M{"$elemMatch": bson.M{"$eq": notificationType}}}
	cursor, err := tokenCollection.Find(ctx, filter)
	if err != nil {
		log.Printf("Erro ao buscar tokens: %v\n", err)
		return err
	}

	var tokens []models.DeviceToken
	if err := cursor.All(ctx, &tokens); err != nil {
		log.Printf("Erro ao processar tokens: %v\n", err)
		return err
	}

	var tokenStrings []string
	for _, token := range tokens {
		tokenStrings = append(tokenStrings, token.DeviceToken)
	}

	err = SendPushNotification(tokenStrings, "Nova Notificação", text)
	if err != nil {
		log.Printf("Erro ao enviar notificações push: %v\n", err)
		return err
	}

	log.Println("Notificação criada e push enviado com sucesso")
	return nil
}

func SendPushNotification(tokens []string, title string, body string) error {
	app, err := InitFirebaseApp()
	if err != nil {
		return err
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Printf("Erro ao inicializar cliente de mensagens: %v\n", err)
		return err
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	response, err := client.SendEachForMulticast(context.Background(), message)
	if err != nil {
		log.Printf("Erro ao enviar mensagem: %v\n", err)
		return err
	}

	log.Printf("Mensagens enviadas com sucesso: %d\n", response.SuccessCount)
	return nil
}

func SendPrivateNotification(userId string, title string, body string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var token models.DeviceToken
	err := tokenCollection.FindOne(ctx, bson.M{"userId": userId}).Decode(&token)
	if err != nil {
		log.Printf("Erro ao buscar token privado para UID %s: %v\n", userId, err)
		return err
	}

	err = SendPushNotification([]string{token.DeviceToken}, title, body)
	if err != nil {
		log.Printf("Erro ao enviar notificação privada para UID %s: %v\n", userId, err)
		return err
	}

	log.Printf("Notificação privada enviada para UID %s com sucesso\n", userId)
	return nil
}

func isUID(value string) bool {
	return len(value) == 24
}
