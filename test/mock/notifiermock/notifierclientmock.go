package notifiermock

import (
	"context"
	openapi "github.com/Interhyp/metadata-service/api"
)

type NotifierClientMock struct {
	SentNotifications []openapi.Notification
}

func (n *NotifierClientMock) Setup(clientIdentifier string, url string) error {
	return nil
}

func (n *NotifierClientMock) Send(ctx context.Context, notification openapi.Notification) {
	n.SentNotifications = append(n.SentNotifications, notification)
}

func (n *NotifierClientMock) Reset() {
	n.SentNotifications = make([]openapi.Notification, 0)
}
