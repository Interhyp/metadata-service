package cache

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	"github.com/Interhyp/metadata-service/internal/service/cache/cacheable"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

type Impl struct {
	Configuration librepo.Configuration
	Logging       librepo.Logging
	Timestamp     librepo.Timestamp

	OwnerCache      cacheable.Cacheable
	ServiceCache    cacheable.Cacheable
	RepositoryCache cacheable.Cacheable
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
	// idempotent to allow mocking
	if s.OwnerCache == nil {
		s.OwnerCache = cacheable.New()
	}
	if s.ServiceCache == nil {
		s.ServiceCache = cacheable.New()
	}
	if s.RepositoryCache == nil {
		s.RepositoryCache = cacheable.New()
	}
	return nil
}
