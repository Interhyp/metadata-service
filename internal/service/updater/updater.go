package updater

import (
	"context"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/StephanHCB/go-backend-service-common/web/middleware/requestid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"reflect"
	"sync"
	"time"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Kafka               repository.Kafka
	Notifier            repository.Notifier
	Mapper              service.Mapper
	Cache               service.Cache

	mu  sync.Mutex
	Now func() time.Time

	totalErrorCounter    prometheus.Counter
	metadataErrorCounter prometheus.Counter
	ownerErrorCounter    *prometheus.CounterVec
	serviceErrorCounter  *prometheus.CounterVec
	repoErrorCounter     *prometheus.CounterVec
}

const (
	removeExisting = iota
	updateExisting = iota
	addNew         = iota
)

var (
	TotalErrorCounterName    = "updater_error_count"
	MetadataErrorCounterName = "updater_error_metadata_count"
	OwnerErrorCounterName    = "updater_error_owner_count"
	ServiceErrorCounterName  = "updater_error_service_count"
	RepoErrorCounterName     = "updater_error_repo_count"
)

// --- metrics ---

func (s *Impl) Setup(ctx context.Context) error {
	s.totalErrorCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: TotalErrorCounterName,
			Help: "How many full update cycles failed.",
		},
	)
	prometheus.MustRegister(s.totalErrorCounter)

	s.metadataErrorCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: MetadataErrorCounterName,
			Help: "How many metadata git operations failed.",
		},
	)
	prometheus.MustRegister(s.metadataErrorCounter)

	s.ownerErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: OwnerErrorCounterName,
			Help: "How many owner updates failed, partitioned by owner alias.",
		},
		[]string{"owner_alias"},
	)
	prometheus.MustRegister(s.ownerErrorCounter)

	s.serviceErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: ServiceErrorCounterName,
			Help: "How many service updates failed, partitioned by service name.",
		},
		[]string{"service_name"},
	)
	prometheus.MustRegister(s.serviceErrorCounter)

	s.repoErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: RepoErrorCounterName,
			Help: "How many repository updates failed, partitioned by repository key.",
		},
		[]string{"repo_key"},
	)
	prometheus.MustRegister(s.repoErrorCounter)

	return nil
}

func (s *Impl) StartReceivingEvents(ctx context.Context) error {
	if err := s.Kafka.SubscribeIncoming(ctx, s.kafkaReceiverCallback); err != nil {
		return err
	}

	return s.Kafka.StartReceiveLoop(ctx)
}

type lockType int

const lockKey lockType = 0

func (s *Impl) WithMetadataLock(ctx context.Context, closure func(context.Context) error) error {
	if ctx.Value(lockKey) == nil {
		s.Logging.Logger().Ctx(ctx).Debug().Print("trying to acquire metadata lock")
		s.mu.Lock()
		s.Logging.Logger().Ctx(ctx).Info().Print("metadata lock acquired")
		defer func() {
			s.Logging.Logger().Ctx(ctx).Debug().Print("trying to release metadata lock")
			s.mu.Unlock()
			s.Logging.Logger().Ctx(ctx).Info().Print("metadata lock released")
		}()

		subCtx := context.WithValue(ctx, lockKey, true)
		err := closure(subCtx)
		return err
	} else {
		s.Logging.Logger().Ctx(ctx).Info().Print("thread already holds metadata lock")
		// we already have the lock (because our context says so)
		return closure(ctx)
	}
}

func (s *Impl) PerformFullUpdate(ctx context.Context) error {
	return s.WithMetadataLock(ctx, func(subCtx context.Context) error {
		if _, err := s.updateMetadata(subCtx); err != nil {
			return err
		}

		if err := s.updateOwners(subCtx); err != nil {
			return err
		}

		if err := s.updateServices(subCtx); err != nil {
			return err
		}

		if err := s.updateRepositories(subCtx); err != nil {
			return err
		}

		return nil
	})
}

func (s *Impl) PerformFullUpdateWithNotifications(ctx context.Context) error {
	return s.WithMetadataLock(ctx, func(subCtx context.Context) error {
		events, err := s.updateMetadata(subCtx)
		if err != nil {
			return err
		}

		if err := s.updateOwners(subCtx); err != nil {
			return err
		}

		if err := s.updateServices(subCtx); err != nil {
			return err
		}

		if err := s.updateRepositories(subCtx); err != nil {
			return err
		}

		for _, event := range events {
			s.fireAndForgetKafkaNotification(subCtx, event)
		}

		return nil
	})
}

func (s *Impl) fireAndForgetKafkaNotification(ctx context.Context, event repository.UpdateEvent) {
	s.Logging.Logger().Ctx(ctx).Debug().Print("preparing to send kafka event")
	err := s.Kafka.Send(ctx, event)
	if err != nil {
		s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Print("failed to send kafka message - continuing anyway")
		// intentionally ignored, kafka is allowed to fail
	}
	s.Logging.Logger().Ctx(ctx).Debug().Print("successfully sent kafka event")
}

func (s *Impl) kafkaReceiverCallback(event repository.UpdateEvent) {
	ctx := context.Background()

	// add custom request id
	requestId := requestid.NewRequestID()
	ctx = context.WithValue(ctx, requestid.RequestIDKey, requestId)

	// add logger
	loggerWithReqId := log.Logger.With().Str("trace.id", requestId).Logger()
	ctx = loggerWithReqId.WithContext(ctx)

	// TODO add timeout
	// seconds := 10
	// ctx, cancel := context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
	// defer cancel()

	err := s.WithMetadataLock(ctx, func(subCtx context.Context) error {
		if s.Mapper.ContainsNewInformation(subCtx, event) {
			s.Logging.Logger().Ctx(subCtx).Info().Printf("received kafka event for new commit hash %s - updating local caches", event.CommitHash)
			return s.PerformFullUpdate(subCtx)
		}
		return nil
	})
	if err != nil {
		// TODO react to timeout and log that it was a timeout
		s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Print("failed to process incoming kafka event")
	}
}

func equalExceptCacheInfo[T openapi.ServiceDto | openapi.OwnerDto | openapi.RepositoryDto](first T, second T) bool {
	clean := func(in *T) T {
		cleaned := new(T)
		*cleaned = *in
		valueOfCleaned := reflect.ValueOf(cleaned).Elem()
		valueOfCleaned.FieldByName("TimeStamp").SetString("")
		valueOfCleaned.FieldByName("CommitHash").SetString("")
		valueOfCleaned.FieldByName("JiraIssue").SetString("")
		return *cleaned
	}
	return reflect.DeepEqual(clean(&first), clean(&second))
}
