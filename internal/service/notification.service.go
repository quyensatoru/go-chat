package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"encoding/json"
	"fmt"
	"log"

	"github.com/SherClockHolmes/webpush-go"
)

var (
	publicKey  = "BBp2a1gwtWptMs5i99yKwrZRJlhoR8MIUtTkddwVV1vR7TvppSoV9dAU5kmILKKlXrgea7He1DNK4gIxpudrQDA"
	privateKey = "OwZ3OOdgjfP9GzZ9C6kOvwp053RSOgedAja2rTbhMjw"
)

type NotificationService interface {
	SubscribeHandler(subscription *Subscription, user *model.User) error
	SendNotification(payload NotificationPayload, user *model.User) error
}

type Subscription struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

type NotificationPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type notificationService struct {
	notifyRepo repository.NotificationRepository
}

func NewNotificationService(notifyRepo repository.NotificationRepository) NotificationService {
	return &notificationService{
		notifyRepo: notifyRepo,
	}
}

var subs []Subscription

func (s *notificationService) SubscribeHandler(subscription *Subscription, user *model.User) error {

	subs = append(subs, *subscription)

	fmt.Println("Đã nhận subscription:", subscription.Endpoint)

	//save notification token to user db
	notify := model.Notification{
		Endpoint: subscription.Endpoint,
		P256dh:   subscription.Keys.P256dh,
		Auth:     subscription.Keys.Auth,
		User:     *user,
	}

	err := s.notifyRepo.CreateOrUpdate(&notify)

	if err != nil {
		return err
	}

	return nil
}

func (s *notificationService) SendNotification(payload NotificationPayload, user *model.User) error {
	notification, err := s.notifyRepo.FindByUserID(user.ID)
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := webpush.SendNotification(jsonBytes, &webpush.Subscription{
		Endpoint: notification.Endpoint,
		Keys: webpush.Keys{
			P256dh: notification.P256dh,
			Auth:   notification.Auth,
		},
	}, &webpush.Options{
		Subscriber:      "mailto:quyenpv020803@gmail.com",
		VAPIDPublicKey:  publicKey,
		VAPIDPrivateKey: privateKey,
		TTL:             60,
	})

	if err != nil {
		log.Println("Push lỗi:", err)
		return err
	}
	defer resp.Body.Close()

	log.Println("Push OK!")
	return nil
}
