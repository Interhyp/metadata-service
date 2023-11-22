package cache

import (
	"context"
	openapi "github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	libcache "github.com/Roshick/go-autumn-synchronisation/pkg/aucache"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"time"
)

var cacheRetention = 30 * 24 * time.Hour

type Impl struct {
	Configuration librepo.Configuration
	Logging       librepo.Logging
	Timestamp     librepo.Timestamp

	OwnerCache      libcache.Cache[openapi.OwnerDto]
	ServiceCache    libcache.Cache[openapi.ServiceDto]
	RepositoryCache libcache.Cache[openapi.RepositoryDto]
	TimestampCache  libcache.Cache[string]
}

func New(
	configuration librepo.Configuration,
	logging librepo.Logging,
	timestamp librepo.Timestamp,
) service.Cache {
	return &Impl{
		Configuration: configuration,
		Logging:       logging,
		Timestamp:     timestamp,
	}
}

func (s *Impl) IsCache() bool {
	return true
}

func (s *Impl) Setup() error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := s.SetupCache(ctx); err != nil {
		s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up business layer cache. BAILING OUT")
		return err
	}

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up cache")
	return nil
}

func (s *Impl) SetupCache(_ context.Context) error {
	// TODO create redis instances based on configuration

	// idempotent to allow mocking
	if s.OwnerCache == nil {
		s.OwnerCache = libcache.NewMemoryCache[openapi.OwnerDto]()
	}
	if s.ServiceCache == nil {
		s.ServiceCache = libcache.NewMemoryCache[openapi.ServiceDto]()
	}
	if s.RepositoryCache == nil {
		s.RepositoryCache = libcache.NewMemoryCache[openapi.RepositoryDto]()
	}
	if s.TimestampCache == nil {
		s.TimestampCache = libcache.NewMemoryCache[string]()
	}
	return nil
}
