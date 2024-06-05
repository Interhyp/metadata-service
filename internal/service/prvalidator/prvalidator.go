package prvalidator

import (
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Timestamp           librepo.Timestamp
	BitBucket           repository.Bitbucket
}

func New(
	configuration librepo.Configuration,
	customConfig config.CustomConfiguration,
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	bitbucket repository.Bitbucket,
) service.PRValidator {
	return &Impl{
		Configuration:       configuration,
		CustomConfiguration: customConfig,
		Logging:             logging,
		Timestamp:           timestamp,
		BitBucket:           bitbucket,
	}
}

func (s *Impl) IsPRValidator() bool {
	return true
}

func (s *Impl) ValidatePullRequest(id uint64, toRef string, fromRef string) error {
	// TODO

	return nil
}
