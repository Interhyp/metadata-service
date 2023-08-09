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

	require.NotNil(t, mockClient.SentNotification)
	sent := *mockClient.SentNotification
	require.Equal(t, name, sent.Name)
	require.Equal(t, event.String(), sent.Event)
	require.Equal(t, payloadType.String(), sent.Type)
	require.Equal(t, payload, sent.Payload)
}
