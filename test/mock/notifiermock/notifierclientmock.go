package notifiermock

import (
	"context"
	"encoding/json"
	"fmt"
	openapi "github.com/Interhyp/metadata-service/api"
)

type NotifierClientMock struct {
	SentNotifications []string
}

func (n *NotifierClientMock) Setup(clientIdentifier string, url string) error {
	return nil
}

func (n *NotifierClientMock) Send(ctx context.Context, notification openapi.Notification) {
	n.SentNotifications = append(n.SentNotifications, n.ToJson(notification))
}

func (n *NotifierClientMock) Reset() {
	n.SentNotifications = make([]string, 0)
}

func (n *NotifierClientMock) ToJson(notification openapi.Notification) string {
	notificationJson, err := json.Marshal(&notification)
	if err != nil {
		notificationJson = []byte(fmt.Sprintf("error: %s", err.Error()))
	}
	return string(notificationJson)
}
