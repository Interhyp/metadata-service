package cache

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/service/cache/cacheable"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"time"
)

type Impl struct {
	Configuration librepo.Configuration
	Logging       librepo.Logging

	OwnerCache      cacheable.Cacheable
	ServiceCache    cacheable.Cacheable
	RepositoryCache cacheable.Cacheable

	Now func() time.Time
}

func (s *Impl) Setup(_ context.Context) error {
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
