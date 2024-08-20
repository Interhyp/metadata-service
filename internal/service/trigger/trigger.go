package trigger

import (
	"context"
	"fmt"
	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/go-backend-service-common/web/middleware/requestid"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"time"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Timestamp           librepo.Timestamp
	Updater             service.Updater

	LoggingCtx context.Context
	Cron       *cron.Cron

	SkipStart bool
}

func New(
	configuration librepo.Configuration,
	customConfig config.CustomConfiguration,
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	updater service.Updater,
) service.Trigger {
	return &Impl{
		Configuration:       configuration,
		CustomConfiguration: customConfig,
		Logging:             logging,
		Timestamp:           timestamp,
		Updater:             updater,
	}
}

func (s *Impl) IsTrigger() bool {
	return true
}

func (s *Impl) Setup() error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := s.SetupTriggerCronjob(ctx); err != nil {
		s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up trigger. BAILING OUT")
		return err
	}

	s.Logging.Logger().Ctx(ctx).Info().Print("performing initial cache population...")

	if err := s.PerformWithCancel(context.Background()); err != nil {
		s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("initial cache population failed. BAILING OUT")
		return err
	}

	if !s.SkipStart {
		s.Logging.Logger().Ctx(ctx).Info().Print("starting event receiver...")

		if err := s.Updater.StartReceivingEvents(ctx); err != nil {
			s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to start event receiver. BAILING OUT")
			return err
		}

		s.Logging.Logger().Ctx(ctx).Info().Print("starting cron job...")

		if err := s.StartCronjob(ctx); err != nil {
			s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to start cron job. BAILING OUT")
			return err
		}
	}

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up trigger")
	return nil
}

func (s *Impl) Teardown() {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	s.Logging.Logger().Ctx(ctx).Info().Print("stopping cron job...")

	if err := s.StopCronjob(ctx); err != nil {
		s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to stop cron job. Continuing with teardown.")
		// do NOT abort tear down cycle
		return
	}

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully tore down trigger")
}

// --- implement cron.Logger ---

func (s *Impl) Info(msg string, keysAndValues ...interface{}) {
	logMessage := msg
	for i, kv := range keysAndValues {
		if i%2 == 0 {
			logMessage += fmt.Sprintf(" %v=", kv)
		} else {
			logMessage += fmt.Sprintf("%v", kv)
		}
	}
	s.Logging.Logger().Ctx(s.LoggingCtx).Info().Print("cronjob: " + logMessage)
}

func (s *Impl) Error(err error, msg string, keysAndValues ...interface{}) {
	logMessage := msg
	for i, kv := range keysAndValues {
		if i%2 == 0 {
			logMessage += fmt.Sprintf(" %v=", kv)
		} else {
			logMessage += fmt.Sprintf("%v", kv)
		}
	}
	// no you don't get to log as Error severity!
	s.Logging.Logger().Ctx(s.LoggingCtx).Warn().WithErr(err).Print("cronjob: " + logMessage)
}

// --- cron job ---

func (s *Impl) SetupTriggerCronjob(ctx context.Context) error {
	s.LoggingCtx = ctx

	s.Cron = cron.New(
		cron.WithLogger(s),
		cron.WithChain(
			cron.SkipIfStillRunning(s),
		),
	)

	cronSpec := fmt.Sprintf("*/%s * * * *", s.CustomConfiguration.UpdateJobIntervalCronPart())
	_, err := s.Cron.AddFunc(cronSpec, func() { _ = s.PerformWithCancel(context.Background()) })
	return err
}

func (s *Impl) StartCronjob(_ context.Context) error {
	s.Cron.Start()
	return nil
}

func (s *Impl) StopCronjob(_ context.Context) error {
	stillRunningCtx := s.Cron.Stop()
	select {
	case <-stillRunningCtx.Done():
		// all jobs have ended
		break
	case <-time.After(30 * time.Second):
		// grace period end
	}

	return nil
}

func (s *Impl) PerformWithCancel(ctx context.Context) error {
	// add custom request id
	requestId := requestid.NewRequestID()
	ctx = context.WithValue(ctx, requestid.RequestIDKey, requestId)

	// add logger
	loggerWithReqId := log.Logger.With().Str("trace.id", requestId).Logger()
	ctx = loggerWithReqId.WithContext(ctx)

	// add timeout
	seconds := s.CustomConfiguration.UpdateJobTimeoutSeconds()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
	defer cancel()

	started := time.Now()

	s.Logging.Logger().Ctx(ctx).Info().Print("starting update")
	err := s.Updater.PerformFullUpdate(ctx)
	tookMs := time.Now().Sub(started).Milliseconds()
	if err != nil {
		s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("finished periodic update with errors (%d ms runtime) - not all information was updated", tookMs)
	} else {
		s.Logging.Logger().Ctx(ctx).Info().Printf("finished update OK (%d ms runtime)", tookMs)
	}
	return err
}
