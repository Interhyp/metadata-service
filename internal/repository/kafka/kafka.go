package kafka

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Roshick/go-autumn-kafka/pkg/kafka"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/rcrowley/go-metrics"
)
import _ "github.com/go-git/go-git/v5"

const MetadataChangeEventsTopicKey = "metadata-change-events"

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	HostIP              repository.HostIP

	Callback      repository.ReceiverCallback
	KafkaProducer *kafka.SyncProducer[repository.UpdateEvent]
	KafkaConsumer *kafka.Consumer[repository.UpdateEvent]
}

func New(
	configuration librepo.Configuration,
	customConfig config.CustomConfiguration,
	logging librepo.Logging,
	hostIP repository.HostIP,
) repository.Kafka {
	return &Impl{
		Callback: func(_ repository.UpdateEvent) {},

		Configuration:       configuration,
		CustomConfiguration: customConfig,
		Logging:             logging,
		HostIP:              hostIP,
	}
}

func (r *Impl) IsKafka() bool {
	return true
}

func (r *Impl) Setup() error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := r.ConnectProducer(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up kafka producer connection. BAILING OUT")
		return err
	}

	r.Logging.Logger().Ctx(ctx).Info().Print("successfully set up kafka producer")
	return nil
}

func (r *Impl) Teardown() {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := r.Disconnect(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to tear down kafka connection(s). Continuing anyway.")
	}
}

func (r *Impl) SubscribeIncoming(ctx context.Context, callback repository.ReceiverCallback) error {
	r.Logging.Logger().Ctx(ctx).Info().Print("accepted kafka subscription callback")
	r.Callback = callback
	return nil
}

func (r *Impl) Send(ctx context.Context, event repository.UpdateEvent) error {
	if r.KafkaProducer == nil {
		return nil
	}

	asyncCtx := context.WithoutCancel(ctx)
	asyncCtx, cancel := context.WithTimeout(asyncCtx, 60*time.Second)
	go func() {
		defer cancel()

		if err := r.KafkaProducer.Produce(asyncCtx, nil, &event); err != nil {
			aulogging.Logger.Ctx(asyncCtx).Warn().WithErr(err).Printf("failed to send event")
		}
	}()
	return nil
}

func (r *Impl) topicConfig(ctx context.Context) (*kafka.TopicConfig, error) {
	if r.CustomConfiguration.Kafka() != nil {
		if topicConfig, ok := r.CustomConfiguration.Kafka().TopicConfigs()[MetadataChangeEventsTopicKey]; ok {
			if topicConfig.Password == "" {
				r.Logging.Logger().Ctx(ctx).Warn().Print("kafka configuration present but password is missing")
				return nil, errors.New("kafka configuration present but got empty password from vault")
			}
			return &topicConfig, nil
		}
	}
	r.Logging.Logger().Ctx(ctx).Info().Print("NOT connecting to kafka due to missing configuration (ok, feature toggle)")
	return nil, nil
}

func (r *Impl) StartReceiveLoop(ctx context.Context) error {
	topicConfig, err := r.topicConfig(ctx)
	if err != nil {
		return err
	}
	if topicConfig == nil {
		return nil
	}

	if r.Callback == nil {
		return errors.New("cannot start kafka receive loop - no callback configured. This is an implementation error")
	}
	callback := func(ctx context.Context, key *string, event *repository.UpdateEvent, stamp time.Time) error {
		if event == nil {
			aulogging.Logger.Ctx(ctx).Warn().Print("kafka receiver callback received nil event - skipping")
			return nil
		}
		r.Callback(*event)
		return nil
	}

	// group id
	groupId := r.CustomConfiguration.KafkaGroupIdOverride()
	if groupId == "" {
		ip, err := r.HostIP.ObtainLocalIp()
		if err != nil {
			return err
		}

		ipComponents := strings.Split(ip.String(), ".")
		if len(ipComponents) != 4 {
			return errors.New("failed to obtain local non-localhost ip address to use for consumer group, did not get an ipv4 address")
		}

		groupId = fmt.Sprintf("metadata-worker-%s-%s", ipComponents[2], ipComponents[3])
	}
	topicConfig.ConsumerGroup = &groupId
	r.Logging.Logger().Ctx(ctx).Info().Printf("using kafka group id %s for consumer", groupId)

	configPreset := sarama.NewConfig()
	configPreset.Net.TLS.Enable = true
	configPreset.Producer.Compression = sarama.CompressionNone
	configPreset.MetricRegistry = metrics.NewPrefixedChildRegistry(metrics.DefaultRegistry, "sarama.consumer.")
	configPreset.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := kafka.CreateConsumer[repository.UpdateEvent](ctx, *topicConfig, callback, configPreset)
	if err != nil {
		return err
	}
	r.Logging.Logger().Ctx(ctx).Info().Print("successfully connected to kafka as consumer (also started receive loop in background)")

	r.KafkaConsumer = consumer
	return nil
}

func (r *Impl) ConnectProducer(ctx context.Context) error {
	topicConfig, err := r.topicConfig(ctx)
	if err != nil {
		return err
	}
	if topicConfig == nil {
		return nil
	}

	configPreset := sarama.NewConfig()
	configPreset.Net.TLS.Enable = true
	configPreset.Producer.Compression = sarama.CompressionNone
	configPreset.MetricRegistry = metrics.NewPrefixedChildRegistry(metrics.DefaultRegistry, "sarama.producer.")

	producer, err := kafka.CreateSyncProducer[repository.UpdateEvent](ctx, *topicConfig, configPreset)
	if err != nil {
		return err
	}
	r.Logging.Logger().Ctx(ctx).Info().Print("successfully connected to kafka as producer")

	r.KafkaProducer = producer
	return nil
}

func (r *Impl) Disconnect(ctx context.Context) error {
	if r.KafkaConsumer != nil {
		r.KafkaConsumer.Close(ctx)
		r.KafkaConsumer = nil
	}
	if r.KafkaProducer != nil {
		r.KafkaProducer.Close(ctx)
		r.KafkaProducer = nil
	}
	return nil
}
