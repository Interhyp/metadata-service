package server

import (
	"context"
	"errors"
	"fmt"
	libcontroller "github.com/Interhyp/go-backend-service-common/acorns/controller"
	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	libmiddleware "github.com/Interhyp/go-backend-service-common/web/middleware"
	"github.com/Interhyp/go-backend-service-common/web/middleware/requestlogging"
	"github.com/Interhyp/go-backend-service-common/web/middleware/security"
	"github.com/Interhyp/metadata-service/internal/acorn/application"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/controller"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	"github.com/go-chi/chi/v5"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
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

func New(
	configuration librepo.Configuration,
	customConfiguration config.CustomConfiguration,
	logging librepo.Logging,
	identityProvider repository.IdentityProvider,
	healthCtl libcontroller.HealthController,
	swaggerCtl libcontroller.SwaggerController,
	ownerCtl controller.OwnerController,
	serviceCtl controller.ServiceController,
	repositoryCtl controller.RepositoryController,
	webhookCtl controller.WebhookController,
) application.Server {
	return &Impl{
		Configuration:       configuration,
		CustomConfiguration: customConfiguration,
		Logging:             logging,
		IdentityProvider:    identityProvider,
		HealthCtl:           healthCtl,
		SwaggerCtl:          swaggerCtl,
		OwnerCtl:            ownerCtl,
		ServiceCtl:          serviceCtl,
		RepositoryCtl:       repositoryCtl,
		WebhookCtl:          webhookCtl,

		RequestTimeoutSeconds:     60,
		ServerWriteTimeoutSeconds: 60,
		ServerIdleTimeoutSeconds:  60,
		ServerReadTimeoutSeconds:  60,
	}
}

func (s *Impl) IsServer() bool {
	return true
}

func (s *Impl) Setup() error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	s.WireUp(ctx)

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up primary web layer")
	return nil
}

func (s *Impl) WireUp(ctx context.Context) {
	if s.Router == nil {
		s.Logging.Logger().Ctx(ctx).Info().Print("creating router and setting up filter chain")
		s.Router = chi.NewRouter()

		keysetPEM := s.IdentityProvider.GetKeySet(ctx)

		options := libmiddleware.MiddlewareStackOptions{
			ElasticApmEnabled:          s.CustomConfiguration.ElasticApmEnabled(),
			PlainLogging:               s.Configuration.PlainLogging(),
			CorsAllowOrigin:            "*", // CORS ok for unauthorized requests
			RequestTimeoutSeconds:      s.RequestTimeoutSeconds,
			HasJwtIdTokenAuthorization: true,
			JwtPublicKeyPEMs:           keysetPEM,
			HasBasicAuthAuthorization:  true,
			BasicAuthUsername:          s.CustomConfiguration.BasicAuthUsername(),
			BasicAuthPassword:          s.CustomConfiguration.BasicAuthPassword(),
			BasicAuthClaims: security.CustomClaims{
				Name:   s.CustomConfiguration.GitCommitterName(),
				Email:  s.CustomConfiguration.GitCommitterEmail(),
				Groups: strings.Fields(s.CustomConfiguration.AuthGroupWrite()),
			},
			AllowUnauthorized: []string{
				// public api endpoints
				"GET /rest/api/v1/owners.*",
				"GET /rest/api/v1/services.*",
				"GET /rest/api/v1/repositories.*",
				"POST /webhook",
				"POST /webhook/bitbucket",
				// health (provides just up)
				"GET /",
				"GET /health",
				"GET /management/health",
				// openapi
				"GET /openapi-v3-spec.yaml",
				"GET /v3/api-docs",
				"GET /swagger-ui.*",
			},
			RequestLoggingOptions: requestlogging.Options{ExcludeLogging: []string{
				"GET / 200",
				"GET /health 200",
				"GET /management/health 200",
			}},
		}

		err := libmiddleware.SetupStandardMiddlewareStack(ctx, s.Router, options)
		if err != nil {
			aulogging.Logger.Ctx(ctx).Fatal().WithErr(err).Printf("failed to set up middleware stack - BAILING OUT: %s", err.Error())
		}
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
