package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/config"
	"github.com/Interhyp/metadata-service/acorns/controller"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/Interhyp/metadata-service/internal/web/middleware"
	"github.com/Interhyp/metadata-service/internal/web/middleware/jwt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	"github.com/StephanHCB/go-autumn-logging-zerolog/loggermiddleware"
	auapmlogging "github.com/StephanHCB/go-autumn-restclient-apm/implementation/logging"
	libcontroller "github.com/StephanHCB/go-backend-service-common/acorns/controller"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/StephanHCB/go-backend-service-common/web/middleware/corsheader"
	"github.com/StephanHCB/go-backend-service-common/web/middleware/requestid"
	"github.com/StephanHCB/go-backend-service-common/web/middleware/requestidinresponse"
	"github.com/StephanHCB/go-backend-service-common/web/middleware/requestlogging"
	"github.com/StephanHCB/go-backend-service-common/web/middleware/requestmetrics"
	"github.com/StephanHCB/go-backend-service-common/web/middleware/timeout"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.elastic.co/apm/module/apmchiv5/v2"
	"go.elastic.co/apm/v2"
	"go.elastic.co/apm/v2/transport"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Impl struct {
	Logging             librepo.Logging
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	IdentityProvider    repository.IdentityProvider
	HealthCtl           libcontroller.HealthController
	SwaggerCtl          libcontroller.SwaggerController
	OwnerCtl            controller.OwnerController
	ServiceCtl          controller.ServiceController
	RepositoryCtl       controller.RepositoryController
	WebhookCtl          controller.WebhookController

	Router chi.Router

	RequestTimeoutSeconds     int
	ServerReadTimeoutSeconds  int
	ServerWriteTimeoutSeconds int
	ServerIdleTimeoutSeconds  int
}

func (s *Impl) WireUp(ctx context.Context) {
	if s.Router == nil {
		s.Logging.Logger().Ctx(ctx).Info().Print("creating router and setting up filter chain")
		s.Router = chi.NewRouter()
		s.Router.Use(middleware.ConstructContextCancellationLoggerMiddleware("Top"))

		loggermiddleware.MethodFieldName = "http.request.method"
		loggermiddleware.PathFieldName = "url.path"

		// auzerolog.RequestIdFieldName changes depending on Json or Plaintext logging
		// auzerolog.RequestIdFieldName is the name used in the logger format
		// loggermiddleware.RequestIdFieldName is the name used to build the request logger
		// if they do not match, the default request id shows in the logs
		loggermiddleware.RequestIdFieldName = auzerolog.RequestIdFieldName

		s.Router.Use(requestid.RequestID)

		s.configureApmMiddleware(ctx)

		s.Router.Use(middleware.ConstructContextCancellationLoggerMiddleware("ElasticApm"))

		// build a request specific logger (includes request id and some fields) and add it to the request context
		s.Router.Use(loggermiddleware.AddZerologLoggerToContext)
		s.Router.Use(middleware.ConstructContextCancellationLoggerMiddleware("AddZerologLoggerToContext"))

		// request logging
		requestlogging.Setup()
		s.Router.Use(chimiddleware.Logger)
		s.Router.Use(middleware.ConstructContextCancellationLoggerMiddleware("Logger"))

		// trap panics in requests and log stack trace
		s.Router.Use(middleware.PanicRecoverer)
		s.Router.Use(middleware.ConstructContextCancellationLoggerMiddleware("PanicRecoverer"))

		// add request id to response, so it can be found in header
		s.Router.Use(requestidinresponse.AddRequestIdHeaderToResponse)
		s.Router.Use(middleware.ConstructContextCancellationLoggerMiddleware("AddRequestIdHeaderToResponse"))

		s.Router.Use(corsheader.CorsHandling)
		s.Router.Use(middleware.ConstructContextCancellationLoggerMiddleware("CorsHandling"))

		requestmetrics.Setup()
		s.Router.Use(requestmetrics.RecordRequestMetrics)
		s.Router.Use(middleware.ConstructContextCancellationLoggerMiddleware("RecordRequestMetrics"))

		_ = jwt.Setup(s.IdentityProvider.GetKeySet(ctx), s.CustomConfiguration)
		s.Router.Use(jwt.JwtValidator)
		s.Router.Use(middleware.ConstructContextCancellationLoggerMiddleware("JwtValidator"))

		timeout.RequestTimeoutSeconds = s.RequestTimeoutSeconds
		s.Router.Use(timeout.AddRequestTimeout)
		s.Router.Use(middleware.ConstructContextCancellationLoggerMiddleware("AddRequestTimeout"))
	}

	s.HealthCtl.WireUp(ctx, s.Router)
	s.SwaggerCtl.WireUp(ctx, s.Router)
	s.OwnerCtl.WireUp(ctx, s.Router)
	s.ServiceCtl.WireUp(ctx, s.Router)
	s.RepositoryCtl.WireUp(ctx, s.Router)
	s.WebhookCtl.WireUp(ctx, s.Router)
}

func (s *Impl) configureApmMiddleware(ctx context.Context) {
	// add apm middleware, because we rely on having a trace context in the context for trace logging and trace propagation to work.
	// Maybe this should be wrapped in the go-autumn-restclient-apm
	if !s.CustomConfiguration.ElasticApmEnabled() {
		// if apm is not configured, we use a discardTracer that does not send any traces
		discardTracer, err := apm.NewTracerOptions(apm.TracerOptions{Transport: transport.Discard})
		if err == nil {
			s.Logging.Logger().Ctx(ctx).Warn().Print("use discard tracer because Elastic APM is not configured")
			// Set defaultTracer as is also used when starting independent transactions (see scheduler)
			//
			apm.SetDefaultTracer(discardTracer)
		}
		// if there was an error creating the discardTracer we stick with the defaultTracer as a crude backup.
		// The default tracer sends its traces to localhost if it is not configured.
	}

	s.Router.Use(apmchiv5.Middleware())

	if s.Configuration.PlainLogging() {
		// set requestIdRetriever to see trace ids in plain logging.
		// skipping it for json logging, because we do not want to interfere with the existing request id middleware
		// and its logging. See else case

		aulogging.RequestIdRetriever = auapmlogging.ExtractTraceId
	} else {
		// we add all the apm specific json log fields as custom fields for json logging
		loggermiddleware.AddCustomJsonLogField(auapmlogging.TraceIdLogFieldName, func(r *http.Request) string {
			return auapmlogging.ExtractTraceId(r.Context())
		})
		loggermiddleware.AddCustomJsonLogField(auapmlogging.TransactionIdLogFieldName, func(r *http.Request) string {
			return auapmlogging.ExtractTransactionId(r.Context())
		})
		loggermiddleware.AddCustomJsonLogField(auapmlogging.SpanIdLogFieldName, func(r *http.Request) string {
			return auapmlogging.ExtractSpanId(r.Context())
		})
	}
}

func (s *Impl) NewServer(ctx context.Context, address string, router http.Handler) *http.Server {
	return &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  time.Duration(s.ServerReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(s.ServerWriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(s.ServerIdleTimeoutSeconds) * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
}

func (s *Impl) CreateMainServer(ctx context.Context) *http.Server {
	address := fmt.Sprintf("%s:%d", s.Configuration.ServerAddress(), s.Configuration.ServerPort())
	s.Logging.Logger().Ctx(ctx).Info().Print("creating primary http server on " + address)
	return s.NewServer(ctx, address, s.Router)
}

func (s *Impl) StartForegroundMainServer(ctx context.Context, srv *http.Server) error {
	s.Logging.Logger().Ctx(ctx).Info().Print("starting primary http server")
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("Fatal error while starting web server: %s\n", err)
	}
	s.Logging.Logger().Ctx(ctx).Info().Print("primary http server has shut down")
	return nil
}

func (s *Impl) Run() error {
	ctxLog := auzerolog.AddLoggerToCtx(context.Background())
	ctx, cancel := context.WithCancel(ctxLog)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	srvMain := s.CreateMainServer(ctx)
	srvMetrics := s.CreateMetricsServer(ctx)

	go func() {
		<-sig // wait for signal notification
		defer cancel()
		s.Logging.Logger().Ctx(ctx).Debug().Print("stopping services now")

		tCtx, tcancel := context.WithTimeout(ctx, 30*time.Second)
		defer tcancel()

		if err := srvMain.Shutdown(tCtx); err != nil {
			s.Logging.Logger().NoCtx().Error().WithErr(err).Printf("failed to shut down primary http server gracefully within 30 seconds: %s", err.Error())
			// this is not perfect, but we need to terminate the entire process because we've trapped sigterm
			os.Exit(3)
		}
		if srvMetrics != nil {
			if err := srvMetrics.Shutdown(tCtx); err != nil {
				s.Logging.Logger().NoCtx().Error().WithErr(err).Printf("failed to shut down metrics http server gracefully within 30 seconds: %s", err.Error())
				// this is not perfect, but we need to terminate the entire process because we've trapped sigterm
				os.Exit(3)
			}
		}
	}()

	s.StartMetricsServerAsyncTerminatesOnError(ctx, srvMetrics)
	if err := s.StartForegroundMainServer(ctx, srvMain); err != nil {
		s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to start foreground web server. BAILING OUT")
		return err
	}

	s.Logging.Logger().Ctx(ctx).Info().Print("application finished successfully")
	return nil
}
