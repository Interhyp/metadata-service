package notifier

import (
	"context"
	"fmt"
	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/go-backend-service-common/web/util/contexthelper"
	openapi "github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	notifierclient "github.com/Interhyp/metadata-service/internal/repository/notifier/client/notifier"
	"github.com/Interhyp/metadata-service/internal/types"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	"time"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging

	Clients   map[string]notifierclient.NotifierClient
	SkipAsync bool // for tests
}

func New(
	configuration librepo.Configuration,
	customConfig config.CustomConfiguration,
	logging librepo.Logging,
) repository.Notifier {
	return &Impl{
		Configuration:       configuration,
		CustomConfiguration: customConfig,
		Logging:             logging,
	}
}

func (r *Impl) IsNotifier() bool {
	return true
}

func (r *Impl) Setup() error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := r.SetupNotifier(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up notifier client. BAILING OUT")
		return err
	}

	r.Logging.Logger().Ctx(ctx).Info().Print("successfully set up notifier")
	return nil
}

const (
	webhookContextTimeout = 5 * time.Minute
)

var (
	_ repository.Notifier = (*Impl)(nil)
)

func AsPayload[T openapi.OwnerDto | openapi.ServiceDto | openapi.RepositoryDto](dto T) openapi.NotificationPayload {
	switch cast := any(dto).(type) {
	case openapi.OwnerDto:
		return openapi.NotificationPayload{
			Owner: &cast,
		}
	case openapi.ServiceDto:
		return openapi.NotificationPayload{
			Service: &cast,
		}
	case openapi.RepositoryDto:
		return openapi.NotificationPayload{
			Repository: &cast,
		}
	default:
		return openapi.NotificationPayload{}
	}
}

func (r *Impl) SetupNotifier(ctx context.Context) error {
	r.Logging.Logger().Ctx(ctx).Info().Print("setting up notifier clients")

	r.Clients = make(map[string]notifierclient.NotifierClient)
	for clientIdentifier, consumerConfig := range r.CustomConfiguration.NotificationConsumerConfigs() {
		client := notifierclient.New(r.Logging, r.CustomConfiguration)

		err := client.Setup(clientIdentifier, consumerConfig.ConsumerURL)
		if err != nil {
			return err
		}
		r.Clients[clientIdentifier] = client
	}

	return nil
}

func (r *Impl) PublishCreation(ctx context.Context, name string, payload openapi.NotificationPayload) error {
	notificationType := determineType(payload)
	if notificationType == nil {
		return fmt.Errorf("unable to determine payload type")
	}
	r.publish(ctx, name, types.CreatedEvent, *notificationType, &payload)
	return nil
}

func (r *Impl) PublishModification(ctx context.Context, name string, payload openapi.NotificationPayload) error {
	notificationType := determineType(payload)
	if notificationType == nil {
		return fmt.Errorf("unable to determine payload type")
	}
	r.publish(ctx, name, types.ModifiedEvent, *notificationType, &payload)
	return nil
}

func (r *Impl) PublishDeletion(ctx context.Context, name string, payloadType types.NotificationPayloadType) {
	r.publish(ctx, name, types.DeletedEvent, payloadType, nil)
}

func determineType(payload openapi.NotificationPayload) *types.NotificationPayloadType {
	owner := payload.Owner
	service := payload.Service
	repo := payload.Repository

	if owner != nil && service == nil && repo == nil {
		payloadType := types.OwnerPayload
		return &payloadType
	}
	if owner == nil && service != nil && repo == nil {
		payloadType := types.ServicePayload
		return &payloadType
	}
	if owner == nil && service == nil && repo != nil {
		payloadType := types.RepositoryPayload
		return &payloadType
	}
	return nil
}

func (r *Impl) publish(ctx context.Context,
	name string, event types.NotificationEventType,
	payloadType types.NotificationPayloadType,
	payload *openapi.NotificationPayload,
) {
	for identifier, consumerConfig := range r.CustomConfiguration.NotificationConsumerConfigs() {
		if _, ok := consumerConfig.Subscribed[payloadType][event]; ok {
			notification := openapi.Notification{
				Name:    name,
				Event:   event.String(),
				Type:    payloadType.String(),
				Payload: payload,
			}
			client := r.Clients[identifier]
			if r.SkipAsync {
				client.Send(ctx, notification)
			} else {
				copyCtx, cancel := contexthelper.AsyncCopyRequestContext(ctx, "notifyConsumer", "outgoingWebhook")
				asyncCtx, timeoutCtxCancel := context.WithTimeout(copyCtx, webhookContextTimeout)

				go func(withClient notifierclient.NotifierClient) {
					defer cancel()
					defer timeoutCtxCancel()
					defer r.recoverPanic(asyncCtx)
					withClient.Send(asyncCtx, notification)
				}(client)
			}
		}
	}
}

func (r *Impl) recoverPanic(ctx context.Context) {
	if recov := recover(); recov != nil {
		err, ok := recov.(error)
		if !ok {
			err = fmt.Errorf("%v", recov)
		}

		r.Logging.Logger().Ctx(ctx).Error().Printf(err.Error())
	}
}
