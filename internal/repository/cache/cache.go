package cache

import (
	"context"
	openapi "github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	libcache "github.com/Roshick/go-autumn-synchronisation/pkg/cache"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"time"
)

var cacheRetention = 30 * 24 * time.Hour

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Timestamp           librepo.Timestamp

	OwnerCache      libcache.Cache[openapi.OwnerDto]
	ServiceCache    libcache.Cache[openapi.ServiceDto]
	RepositoryCache libcache.Cache[openapi.RepositoryDto]
	TimestampCache  libcache.Cache[string]
}

func New(
	configuration librepo.Configuration,
	customConfig config.CustomConfiguration,
	logging librepo.Logging,
	timestamp librepo.Timestamp,
) repository.Cache {
	return &Impl{
		Configuration:       configuration,
		CustomConfiguration: customConfig,
		Logging:             logging,
		Timestamp:           timestamp,
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

const (
	ownerKeyPrefix      = "v1-owner"
	serviceKeyPrefix    = "v1-service"
	repositoryKeyPrefix = "v1-repository"
	timestampKeyPrefix  = "v1-timestamp"
)

func (s *Impl) SetupCache(ctx context.Context) error {
	redisUrl := s.CustomConfiguration.RedisUrl()
	if redisUrl == "" {
		s.Logging.Logger().Ctx(ctx).Info().Print("using in-memory cache")
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
	} else {
		s.Logging.Logger().Ctx(ctx).Info().Printf("using redis at %s", redisUrl)
		redisPassword := s.CustomConfiguration.RedisUrl()
		if s.OwnerCache == nil {
			cache, err := libcache.NewRedisCache[openapi.OwnerDto](redisUrl, redisPassword, ownerKeyPrefix)
			if err != nil {
				return err
			}
			s.OwnerCache = cache
		}
		if s.ServiceCache == nil {
			cache, err := libcache.NewRedisCache[openapi.ServiceDto](redisUrl, redisPassword, serviceKeyPrefix)
			if err != nil {
				return err
			}
			s.ServiceCache = cache
		}
		if s.RepositoryCache == nil {
			cache, err := libcache.NewRedisCache[openapi.RepositoryDto](redisUrl, redisPassword, repositoryKeyPrefix)
			if err != nil {
				return err
			}
			s.RepositoryCache = cache
		}
		if s.TimestampCache == nil {
			cache, err := libcache.NewRedisCache[string](redisUrl, redisPassword, timestampKeyPrefix)
			if err != nil {
				return err
			}
			s.TimestampCache = cache
		}
	}
	return nil
}
