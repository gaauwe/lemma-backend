package notification

import (
	"errors"
	"log"

	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

type Data struct {
	url string
}

var Client *apns2.Client

func SetupClient() error {
	authKey, err := token.AuthKeyFromFile("./AuthKey.p8")
	if err != nil {
		return err
	}

	token := &token.Token{
		AuthKey: authKey,
		KeyID:   config.Get().Apn.KeyId,
		TeamID:  config.Get().Apn.TeamId,
	}

	Client = apns2.NewTokenClient(token).Development()
	return nil
}

func SendNotification(title string, body string, image string, count int64, url string, user *database.User) {
	// Generate notification body data.
	data := make(map[string]interface{})
	data["url"] = url
	data["username"] = user.Username

	// Generate notification payload.
	payload := payload.NewPayload().AlertTitle(title).AlertBody(body).Sound("default").MutableContent().Custom("image_url", image).Custom("body", data)

	if count > 0 {
		payload.Badge(int(count))
	}

	// Prepare notification.
	notification := &apns2.Notification{}
	notification.DeviceToken = user.DeviceToken
	notification.Topic = config.Get().Apn.Topic
	notification.Payload = payload

	// Send notification.
	res, err := Client.Push(notification)
	if err != nil {
		log.Println("Failed to send notification: ", err)
	}

	if res.Sent() {
		log.Println("Notification Sent: ", res.ApnsID)
	} else {
		log.Println("Notification Not Sent: ", res.StatusCode, res.ApnsID, res.Reason)
	}
}

// This sends a silent notification to the device that tries to register, this way we can verify if it's a valid registration or a bad actor.
func SendRegistrationNotification(token string) error {
	payload := payload.NewPayload().ContentAvailable()

	notification := &apns2.Notification{}
	notification.DeviceToken = token
	notification.Topic = config.Get().Apn.Topic
	notification.Payload = payload

	res, err := Client.Push(notification)
	if err != nil {
		return err
	}

	if res.Sent() {
		return nil
	} else {
		return errors.New("Failed to sent registration notification")
	}
}
