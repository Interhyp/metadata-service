package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/controller"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/Interhyp/metadata-service/web/middleware"
	"github.com/Interhyp/metadata-service/web/middleware/jwt"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	"github.com/StephanHCB/go-autumn-logging-zerolog/loggermiddleware"
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
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Impl struct {
	Logging          librepo.Logging
	Configuration    librepo.Configuration
	Vault            repository.Vault
	IdentityProvider repository.IdentityProvider
	HealthCtl        libcontroller.HealthController
	SwaggerCtl       libcontroller.SwaggerController
	OwnerCtl         controller.OwnerController
	ServiceCtl       controller.ServiceController
	RepositoryCtl    controller.RepositoryController
	WebhookCtl       controller.WebhookController

	Router chi.Router
}

func (s *Impl) WireUp(ctx context.Context) {
	if s.Router == nil {
		s.Logging.Logger().Ctx(ctx).Info().Print("creating router and setting up filter chain")
		s.Router = chi.NewRouter()

		// generate request id (or read from request header if present) and add it to request context
		requestid.RequestIDHeader = "X-B3-TraceId"
		s.Router.Use(requestid.RequestID)

		loggermiddleware.RequestIdFieldName = "trace.id"
		loggermiddleware.MethodFieldName = "http.request.method"
		loggermiddleware.PathFieldName = "url.path"
		// build a request specific logger (includes request id and some fields) and add it to the request context
		s.Router.Use(loggermiddleware.AddZerologLoggerToContext)

		// request logging
		requestlogging.Setup()
		s.Router.Use(chimiddleware.Logger)

		// trap panics in requests and log stack trace
		s.Router.Use(middleware.PanicRecoverer)

		// add request id to response, so it can be found in header
		s.Router.Use(requestidinresponse.AddRequestIdHeaderToResponse)

		s.Router.Use(corsheader.CorsHandling)

		requestmetrics.Setup()
		s.Router.Use(requestmetrics.RecordRequestMetrics)

		_ = jwt.Setup(s.IdentityProvider.GetKeySet(ctx), s.Vault.BasicAuthUsername(), s.Vault.BasicAuthPassword())
		s.Router.Use(jwt.JwtValidator)

		s.Router.Use(timeout.AddRequestTimeout)
	}

	s.HealthCtl.WireUp(ctx, s.Router)
	s.SwaggerCtl.WireUp(ctx, s.Router)
	s.OwnerCtl.WireUp(ctx, s.Router)
	s.ServiceCtl.WireUp(ctx, s.Router)
	s.RepositoryCtl.WireUp(ctx, s.Router)
	s.WebhookCtl.WireUp(ctx, s.Router)
}

func (s *Impl) NewServer(ctx context.Context, address string, router http.Handler) *http.Server {
	return &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
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
