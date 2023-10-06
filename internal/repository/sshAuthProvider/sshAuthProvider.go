package sshAuthProvider

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"

	aulogging "github.com/StephanHCB/go-autumn-logging"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type SshAuthProviderImpl struct {
	Configuration librepo.Configuration
	Logging       librepo.Logging

	CustomConfiguration config.CustomConfiguration
}

func New(
	configuration librepo.Configuration,
	customConfig config.CustomConfiguration,
	logging librepo.Logging,
) repository.SshAuthProvider {
	return &SshAuthProviderImpl{
		Configuration:       configuration,
		CustomConfiguration: customConfig,
		Logging:             logging,
	}
}

func (s *SshAuthProviderImpl) IsSshAuthProvider() bool {
	return true
}

func (s *SshAuthProviderImpl) Setup() error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := s.SetupProvider(ctx); err != nil {
		s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up business layer SshAuthProvider. BAILING OUT")
		return err
	}

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up SshAuthProvider service")
	return nil
}

func (s *SshAuthProviderImpl) SetupProvider(ctx context.Context) error {
	return nil
}

// SshAuthProvider for a business method

func (s *SshAuthProviderImpl) ProvideSshAuth(ctx context.Context) (*ssh.PublicKeys, error) {
	return providePublicFromPrivateSshKey(ctx, s.CustomConfiguration.SSHPrivateKey(), s.CustomConfiguration.SSHPrivateKeyPassword())
}

func providePublicFromPrivateSshKey(ctx context.Context, privateKeyData string, privateKeyFilePassword string) (*ssh.PublicKeys, error) {
	result, err := ssh.NewPublicKeys("git", []byte(privateKeyData), privateKeyFilePassword)
	if err != nil {
		warn(ctx, "generation of publickeys failed", err)
		return nil, err
	}
	return result, nil
}

func warn(ctx context.Context, message string, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf(message+": %v", err)
}
