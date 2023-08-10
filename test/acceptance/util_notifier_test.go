package acceptance

import (
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/types"
	"github.com/Interhyp/metadata-service/test/mock/notifiermock"
	"github.com/stretchr/testify/require"
	"testing"
)

func hasSentNotification(t *testing.T, clientIdentifier string, name string, event types.NotificationEventType, payloadType types.NotificationPayloadType, payload *openapi.NotificationPayload) {
	client := notifierImpl.Clients[clientIdentifier]
	mockClient := client.(*notifiermock.NotifierClientMock)
	expected := openapi.Notification{
		Name:    name,
		Event:   event.String(),
		Type:    payloadType.String(),
		Payload: payload,
	}
	notifications := mockClient.SentNotifications
	require.Contains(t, notifications, expected)
}
