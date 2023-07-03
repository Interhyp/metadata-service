package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"net"
	"strings"
	"time"
)
import _ "github.com/go-git/go-git/v5"

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	HostIP              repository.HostIP

	Callback      repository.ReceiverCallback
	KafkaProducer *kgo.Client
	KafkaConsumer *kgo.Client
	KafkaTopic    string
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

	value, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	record := &kgo.Record{Topic: r.KafkaTopic, Value: value}

	// this blocks completely?!?

	//if err := r.KafkaProducer.ProduceSync(ctx, record).FirstErr(); err != nil {
	//	return err
	//}

	// so use async produce (with no guarantee of delivery)

	r.KafkaProducer.Produce(ctx, record, nil)

	return nil
}

func (r *Impl) StartReceiveLoop(ctx context.Context) error {
	r.Logging.Logger().Ctx(ctx).Info().Print("starting receive loop in background")
	go func() {
		myCtx := auzerolog.AddLoggerToCtx(context.Background())
		_ = r.receiveLoop(myCtx)
	}()
	return nil
}

// receiveLoop should terminate when its context is cancelled
//
// it will also terminate on fetch errors
//
// TODO handle fetch errors more than just logging
func (r *Impl) receiveLoop(ctx context.Context) error {
	if r.KafkaConsumer == nil {
		r.Logging.Logger().Ctx(ctx).Info().Print("receive loop cannot start, no kafka client")
		return nil
	}
	for {
		fetches := r.KafkaConsumer.PollFetches(context.Background())
		if fetches.IsClientClosed() {
			r.Logging.Logger().Ctx(ctx).Info().Print("receive loop ending, kafka client was closed")
			return nil
		}
		r.Logging.Logger().Ctx(ctx).Debug().Printf("receive loop found %d fetches", len(fetches))

		var firstError error = nil
		fetches.EachError(func(t string, p int32, err error) {
			if firstError == nil {
				firstError = fmt.Errorf("receive loop fetch error topic %s partition %d: %v", t, p, err)
			}
		})
		if firstError != nil {
			r.Logging.Logger().Ctx(ctx).Error().WithErr(firstError).Print("receive loop terminated abnormally: %v", firstError)
			return firstError
		}

		fetches.EachRecord(func(record *kgo.Record) {
			event := repository.UpdateEvent{}
			err := json.Unmarshal(record.Value, &event)
			if err != nil {
				r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("receive loop json error - ignoring malformed message: %v", err)
			} else {
				r.Logging.Logger().Ctx(ctx).Info().Printf("received kafka message: %v", event)
				r.Callback(event)
			}
		})
	}
}

func (r *Impl) createConsumer(ctx context.Context, seedBrokers string, user string, pass string, topic string) (*kgo.Client, error) {
	groupId := r.CustomConfiguration.KafkaGroupIdOverride()
	if groupId == "" {
		ip, err := r.HostIP.ObtainLocalIp()
		if err != nil {
			return nil, err
		}

		ipComponents := strings.Split(ip.String(), ".")
		if len(ipComponents) != 4 {
			return nil, errors.New("failed to obtain local non-localhost ip address to use for consumer group, did not get an ipv4 address")
		}

		workerNodeId := ipComponents[2]

		groupId = "metadata-worker" + workerNodeId
	}

	r.Logging.Logger().Ctx(ctx).Info().Printf("using kafka group id %s for consumer", groupId)

	tlsDialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 10 * time.Second},
		Config:    &tls.Config{InsecureSkipVerify: true},
	}
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(seedBrokers, ",")...),

		kgo.SASL(scram.Auth{
			User: user,
			Pass: pass,
		}.AsSha256Mechanism()),

		kgo.Dialer(tlsDialer.DialContext),

		kgo.ConsumerGroup(groupId),
		kgo.ConsumeTopics(topic),
		kgo.SessionTimeout(30 * time.Second),
		kgo.WithLogger(r),
	}

	consumer, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

func (r *Impl) createProducer(seedBrokers string, user string, pass string) (*kgo.Client, error) {
	tlsDialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 10 * time.Second},
		Config:    &tls.Config{InsecureSkipVerify: true},
	}
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(seedBrokers, ",")...),

		kgo.SASL(scram.Auth{
			User: user,
			Pass: pass,
		}.AsSha256Mechanism()),

		kgo.Dialer(tlsDialer.DialContext),

		kgo.RequestRetries(2),
		kgo.RetryTimeout(5 * time.Second),
		kgo.WithLogger(r),
	}

	producer, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return producer, nil
}

func (r *Impl) Connect(ctx context.Context) error {
	seedBrokers := r.CustomConfiguration.KafkaSeedBrokers()
	user := r.CustomConfiguration.KafkaUsername()
	pass := r.CustomConfiguration.KafkaPassword()
	topic := r.CustomConfiguration.KafkaTopic()

	if seedBrokers == "" || user == "" || topic == "" {
		r.Logging.Logger().Ctx(ctx).Info().Print("NOT connecting to kafka due to missing configuration (ok, feature toggle)")
		return nil
	}
	if pass == "" {
		r.Logging.Logger().Ctx(ctx).Warn().Print("kafka configuration present but password is missing")
		return errors.New("kafka configuration present but got empty password from vault")
	}

	consumer, err := r.createConsumer(ctx, seedBrokers, user, pass, topic)
	if err != nil {
		return err
	}
	r.Logging.Logger().Ctx(ctx).Info().Print("successfully connected to kafka as consumer")

	producer, err := r.createProducer(seedBrokers, user, pass)
	if err != nil {
		consumer.Close()
		return err
	}
	r.Logging.Logger().Ctx(ctx).Info().Print("successfully connected to kafka as producer")

	r.KafkaConsumer = consumer
	r.KafkaProducer = producer
	r.KafkaTopic = topic
	return nil
}

func (r *Impl) Disconnect(ctx context.Context) error {
	if r.KafkaConsumer != nil {
		r.KafkaConsumer.Close()
		r.KafkaConsumer = nil
	}
	if r.KafkaProducer != nil {
		r.KafkaProducer.Close()
		r.KafkaProducer = nil
	}
	return nil
}

// --- implementing kgo.Logger ---

func (r *Impl) Level() kgo.LogLevel {
	return kgo.LogLevelInfo
}

func (r *Impl) Log(level kgo.LogLevel, msg string, keyvals ...interface{}) {
	switch level {
	case kgo.LogLevelError:
		r.Logging.Logger().NoCtx().Warn().Print("kgo error: " + msg)
		return
	case kgo.LogLevelWarn:
		r.Logging.Logger().NoCtx().Warn().Print("kgo warning: " + msg)
		return
	case kgo.LogLevelInfo:
		r.Logging.Logger().NoCtx().Debug().Print("kgo info: " + msg)
		return
	case kgo.LogLevelDebug:
		r.Logging.Logger().NoCtx().Debug().Print("kgo debug: " + msg)
		return
	default:
		return
	}
}
