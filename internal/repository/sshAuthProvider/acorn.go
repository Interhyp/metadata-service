package sshAuthProvider

import (
	"context"

	"github.com/Interhyp/metadata-service/acorns/config"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &SshAuthProviderImpl{}
}

func (s *SshAuthProviderImpl) IsSshAuthProvider() bool {
	return true
}

func (s SshAuthProviderImpl) AcornName() string {
	return repository.SshAuthProviderAcornName
}

func (s *SshAuthProviderImpl) AssembleAcorn(registry auacornapi.AcornRegistry) error {
	s.Configuration = registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration)
	s.Logging = registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging)

	s.CustomConfiguration = config.Custom(s.Configuration)

	return nil
}

func (s *SshAuthProviderImpl) SetupAcorn(registry auacornapi.AcornRegistry) error {
	if err := registry.SetupAfter(s.Logging.(auacornapi.Acorn)); err != nil {
		return err
	}

	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := s.Setup(ctx); err != nil {
		s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up business layer SshAuthProvider. BAILING OUT")
		return err
	}

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up SshAuthProvider service")
	return nil
}

func (s *SshAuthProviderImpl) TeardownAcorn(_ auacornapi.AcornRegistry) error {
	return nil
}
