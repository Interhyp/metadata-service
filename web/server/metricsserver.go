package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// we use this with a go-routine, as it never returns
func (s *Impl) metricsServerListenAndServe(ctx context.Context, srv *http.Server) {
	s.Logging.Logger().Ctx(ctx).Info().Print("starting metrics http server")
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.Logging.Logger().NoCtx().Error().WithErr(err).Print("failed to start background metrics server. BAILING OUT")
		s.Logging.Logger().NoCtx().Fatal().WithErr(err).Print("error was: " + err.Error())
	}
	s.Logging.Logger().NoCtx().Info().Print("metrics http server has shut down")
}

func (s *Impl) StartMetricsServerAsyncTerminatesOnError(ctx context.Context, srv *http.Server) {
	if srv != nil {
		go s.metricsServerListenAndServe(ctx, srv)
	}
}

func (s *Impl) CreateMetricsServer(ctx context.Context) *http.Server {
	if s.Configuration.MetricsPort() > 0 {
		address := fmt.Sprintf("%s:%d", s.Configuration.ServerAddress(), s.Configuration.MetricsPort())
		s.Logging.Logger().Ctx(ctx).Info().Print("creating metrics http server on " + address)
		metricsServeMux := http.NewServeMux()
		metricsServeMux.Handle("/metrics", promhttp.Handler())
		return s.NewServer(ctx, address, metricsServeMux)
	} else {
		s.Logging.Logger().Ctx(ctx).Info().Print("will not start metrics http server - no metrics port configured")
		return nil
	}
}
