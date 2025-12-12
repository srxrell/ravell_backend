package firebase

import (
	"context"
	"fmt"
	"log"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

var FCMClient *messaging.Client

func InitFirebase() {
    opt := option.WithCredentialsFile("./service_accounts/hhhhh-eda4f-firebase-adminsdk-fbsvc-cb28815df6.json")
    app, err := firebase.NewApp(context.Background(), nil, opt)
    if err != nil {
        log.Fatalf("error initializing firebase app: %v\n", err)
    }

    client, err := app.Messaging(context.Background())
    if err != nil {
        log.Fatalf("error getting messaging client: %v\n", err)
    }

    FCMClient = client
    log.Println("✅ Firebase initialized")
}

type PushRequest struct {
    PlayerIDs []string `json:"player_ids"`
    Title     string   `json:"title"`
    Body      string   `json:"body"`
}

func SendPushNotificationHandler(c *gin.Context) {
    var req PushRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if len(req.PlayerIDs) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No player IDs provided"})
        return
    }

    err := SendPush(req.PlayerIDs, req.Title, req.Body)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Push sent"})
}

func SendPush(playerIDs []string, title, body string) error {
    if FCMClient == nil {
        return fmt.Errorf("FCM client not initialized")
    }

    message := &messaging.MulticastMessage{
        Tokens: playerIDs,
        Notification: &messaging.Notification{
            Title: title,
            Body:  body,
        },
    }

    _, err := FCMClient.SendMulticast(context.Background(), message)
    return err
}
