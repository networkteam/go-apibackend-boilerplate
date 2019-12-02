package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/friendsofgo/errors"

)

type PushNotificationService struct {
	goRushApiUrl string
	HttpClient   http.Client
}

type soundPayload struct {
	Critical int
	Volume   float32
	Name     string
}

type pushNotification struct {
	Tokens   []string    `json:"tokens"`
	Platform int         `json:"platform"`
	Message  string      `json:"message"`
	Topic    *string     `json:"topic,omitempty"`
	Sound    interface{} `json:"sound"`
	Data     interface{} `json:"data"`
}

type PushNotificationInput struct {
	Message      string
	Data         map[string]interface{}
}

func NewPushNotificationService(goRushApiUrl string) *PushNotificationService {
	return &PushNotificationService{
		goRushApiUrl: goRushApiUrl,
		HttpClient:   http.Client{Timeout: time.Second * 10},
	}
}

func (h PushNotificationService) Notify(registration DeviceRegistrationProvider, payload PayloadProvider) error {
	message := payload.GetMessage()
	data := payload.GetData()

	if registration.GetDeviceOS() == "" {
		return errors.New("no device OS set for registration")
	}

	if registration.GetDeviceToken() == "" {
		return errors.New("no device token set for registration")
	}

	notification, err := buildPushNotification(strings.ToLower(registration.GetDeviceOS()), message, data, registration)
	if err != nil {
		return errors.Wrap(err, "could not build push notification")
	}

	err = h.sendPushNotification(notification)
	if err != nil {
		return errors.Wrap(err, "could not send push notification")
	}

	return nil
}

func (h PushNotificationService) sendPushNotification(notification pushNotification) error {
	payload := map[string]interface{}{
		"notifications": []pushNotification{notification},
	}

	jsonValue, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "marshalling notification JSON payload")
	}

	jsonPayload := bytes.NewBuffer(jsonValue)

	response, err := h.HttpClient.Post(fmt.Sprintf("%s/push", h.goRushApiUrl), "application/json", jsonPayload)
	if err != nil {
		return errors.Wrap(err, "posting to Gorush")
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.WithError(err).Error("failed to close response body")
		}
	}()

	if response.Status != "200" {
		return errors.Wrap(err, "Gorush responded with non 200 status code")
	}

	return nil
}

func buildPushNotification(deviceOs string, message string, data interface{}, registration DeviceRegistrationProvider) (notification pushNotification, err error) {
	if deviceOs == "ios" {
		return buildIosPushNotification(message, data, registration), nil
	}

	if deviceOs == "android" {
		return buildAndroidPushNotification(message, data, registration), nil
	}

	return notification, errors.New(fmt.Sprintf("device OS %q not supported", deviceOs))
}

func buildIosPushNotification(message string, data interface{}, registration DeviceRegistrationProvider) (notification pushNotification) {
	var tokens []string

	tokens = append(tokens, registration.GetDeviceToken())

	topic := "mytld.myvendor.myproject"

	sound := soundPayload{
		Critical: 1,
		Name:     "default",
		Volume:   2.0,
	}

	return pushNotification{
		Platform: 1,
		Tokens:   tokens,
		Message:  message,
		Topic:    &topic,
		Sound:    &sound,
		Data:     data,
	}
}

func buildAndroidPushNotification(message string, data interface{}, registration DeviceRegistrationProvider) (notification pushNotification) {
	var tokens []string
	tokens = append(tokens, registration.GetDeviceToken())

	return pushNotification{
		Platform: 2,
		Tokens:   tokens,
		Message:  message,
		Sound:    "default",
		Data:     data,
	}
}
