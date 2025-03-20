package authProvider

import (
	"context"
	"fmt"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	githubclient "github.com/Interhyp/metadata-service/internal/repository/github"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-git/go-git/v5/plumbing/transport"
	ghhttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/gofri/go-github-pagination/githubpagination"
	"net/http"
	"time"

	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/google/go-github/v69/github"
)

type AuthProviderFn func(context.Context) (transport.AuthMethod, error)

type AuthProviderImpl struct {
	Configuration librepo.Configuration
	Logging       librepo.Logging

	CustomConfiguration config.CustomConfiguration

	Github repository.Github

	token          *github.InstallationToken
	authProviderFn AuthProviderFn
}

func New(
	configuration librepo.Configuration,
	customConfig config.CustomConfiguration,
	logging librepo.Logging,
	baseRT http.RoundTripper,
) (repository.AuthProvider, error) {
	jwtTransport, err := ghinstallation.NewAppsTransport(baseRT, customConfig.GithubAppId(), customConfig.GithubAppJwtSigningKeyPEM())
	paginator := githubpagination.NewClient(jwtTransport,
		githubpagination.WithPerPage(100),
		githubpagination.WithMaxNumOfPages(10),
	)
	githubClient := githubclient.New(nil, github.NewClient(paginator))

	return &AuthProviderImpl{
		Configuration:       configuration,
		CustomConfiguration: customConfig,
		Logging:             logging,
		Github:              githubClient,
	}, err
}

func (s *AuthProviderImpl) IsAuthProvider() bool {
	return true
}

func (s *AuthProviderImpl) Setup() error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := s.SetupProvider(ctx); err != nil {
		s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up business layer AuthProvider. BAILING OUT")
		return err
	}

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up AuthProvider service")
	return nil
}

func (s *AuthProviderImpl) SetupProvider(_ context.Context) error {
	s.authProviderFn = s.GetAuth
	return nil
}

func (s *AuthProviderImpl) ProvideAuth(ctx context.Context) transport.AuthMethod {
	auth, _ := s.authProviderFn(ctx)
	s.Logging.Logger().Ctx(ctx).Trace().Print("using basic auth for github")
	return auth
}

// AuthProvider for a business method
func (s *AuthProviderImpl) GetAuth(ctx context.Context) (transport.AuthMethod, error) {
	if s.token == nil || s.token.GetExpiresAt().Before(time.Now().Add(-30*time.Second)) {
		var err error
		aulogging.Logger.Ctx(ctx).Trace().Print("creat new installation token for org")
		s.token, _, err = s.Github.CreateInstallationToken(ctx, s.CustomConfiguration.GithubAppInstallationId())
		if err != nil {
			return nil, fmt.Errorf("failed to create installation token: %w", err)
		}
	}

	return &ghhttp.BasicAuth{
		Username: "x-access-token",
		Password: s.token.GetToken(),
	}, nil
}
