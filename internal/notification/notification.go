package notification

import (
	"log"

	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

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

func SendNotification(title string, body string, image string, count int64) {
	payload := payload.NewPayload().AlertTitle(title).AlertBody(body).MutableContent().Badge(int(count)).Custom("image_url", image)

	notification := &apns2.Notification{}
	notification.DeviceToken = config.Get().Device.DeviceToken
	notification.Topic = config.Get().Device.Topic
	notification.Payload = payload

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
