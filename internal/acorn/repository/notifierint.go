package repository

import (
	"context"
	openapi "github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/types"
)

type Notifier interface {
	IsNotifier() bool

	Setup() error

	SetupNotifier(ctx context.Context) error

	PublishCreation(ctx context.Context, payloadName string, payload openapi.NotificationPayload) error

	PublishModification(ctx context.Context, payloadName string, payload openapi.NotificationPayload) error

	PublishDeletion(ctx context.Context, payloadName string, payloadType types.NotificationPayloadType)
}
