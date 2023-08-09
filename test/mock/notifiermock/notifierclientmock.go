package notifiermock

import (
	"context"
	openapi "github.com/Interhyp/metadata-service/api"
)

type NotifierClientMock struct {
	SentNotification *openapi.Notification
}

func (n *NotifierClientMock) Setup(clientIdentifier string, url string) error {
	return nil
}

func (n *NotifierClientMock) Send(ctx context.Context, notification openapi.Notification) {
	n.SentNotification = &notification
}

func (n *NotifierClientMock) Reset() {
	n.SentNotification = nil
}
