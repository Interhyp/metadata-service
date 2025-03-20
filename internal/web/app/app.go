package app

import (
	libcontroller "github.com/Interhyp/go-backend-service-common/acorns/controller"
	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/go-backend-service-common/repository/logging"
	"github.com/Interhyp/go-backend-service-common/repository/timestamp"
	"github.com/Interhyp/go-backend-service-common/repository/vault"
	"github.com/Interhyp/go-backend-service-common/web/controller/healthctl"
	"github.com/Interhyp/go-backend-service-common/web/controller/swaggerctl"
	"github.com/Interhyp/metadata-service/internal/acorn/application"
	configrepo "github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/controller"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	"github.com/Interhyp/metadata-service/internal/repository/authProvider"
	"github.com/Interhyp/metadata-service/internal/repository/cache"
	"github.com/Interhyp/metadata-service/internal/repository/config"
	"github.com/Interhyp/metadata-service/internal/repository/github"
	"github.com/Interhyp/metadata-service/internal/repository/hostip"
	"github.com/Interhyp/metadata-service/internal/repository/idp"
	"github.com/Interhyp/metadata-service/internal/repository/kafka"
	"github.com/Interhyp/metadata-service/internal/repository/metadata"
	"github.com/Interhyp/metadata-service/internal/repository/notifier"
	"github.com/Interhyp/metadata-service/internal/service/mapper"
	"github.com/Interhyp/metadata-service/internal/service/owners"
	"github.com/Interhyp/metadata-service/internal/service/repositories"
	"github.com/Interhyp/metadata-service/internal/service/services"
	"github.com/Interhyp/metadata-service/internal/service/trigger"
	"github.com/Interhyp/metadata-service/internal/service/updater"
	"github.com/Interhyp/metadata-service/internal/service/validator"
	"github.com/Interhyp/metadata-service/internal/service/webhookshandler"
	"github.com/Interhyp/metadata-service/internal/web/controller/ownerctl"
	"github.com/Interhyp/metadata-service/internal/web/controller/repositoryctl"
	"github.com/Interhyp/metadata-service/internal/web/controller/servicectl"
	"github.com/Interhyp/metadata-service/internal/web/controller/webhookctl"
	"github.com/Interhyp/metadata-service/internal/web/server"
	aurestrecorder "github.com/StephanHCB/go-autumn-restclient/implementation/recorder"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/gofri/go-github-pagination/githubpagination"
	gogithub "github.com/google/go-github/v69/github"
	"net/http"
	"time"
)

type ApplicationImpl struct {
	// repositories (outgoing connectors)
	Config           librepo.Configuration
	CustomConfig     configrepo.CustomConfiguration
	Logging          librepo.Logging
	Vault            librepo.Vault
	Metadata         repository.Metadata
	Kafka            repository.Kafka
	IdentityProvider repository.IdentityProvider
	HostIP           repository.HostIP
	Timestamp        librepo.Timestamp
	AuthProvider     repository.AuthProvider
	Notifier         repository.Notifier
	Cache            repository.Cache
	BaseRT           http.RoundTripper
	Github           repository.Github

	// services (business logic)
	Mapper          service.Mapper
	Trigger         service.Trigger
	Updater         service.Updater
	Validator       service.Validator
	Owners          service.Owners
	Services        service.Services
	Repositories    service.Repositories
	WebhooksHandler service.WebhooksHandler

	// controllers (incoming connectors)
	HealthCtl     libcontroller.HealthController
	SwaggerCtl    libcontroller.SwaggerController
	OwnerCtl      controller.OwnerController
	ServiceCtl    controller.ServiceController
	RepositoryCtl controller.RepositoryController
	WebhookCtl    controller.WebhookController

	// server/web stack
	Server application.Server
}

func New() application.Application {
	return &ApplicationImpl{}
}

func (a *ApplicationImpl) IsApplication() bool {
	return true
}

func (a *ApplicationImpl) Create() error {
	if err := a.ConstructConfigLoggingVaultTimestamp(); err != nil {
		return err
	}

	if err := a.ConstructRepositories(); err != nil {
		return err
	}

	if err := a.ConstructServices(); err != nil {
		return err
	}

	if err := a.ConstructControllers(); err != nil {
		return err
	}

	return nil
}

func (a *ApplicationImpl) Teardown() {
	// reverse order (must ensure correct order yourself, but most components will not have a teardown method)
	a.Trigger.Teardown()
	a.Kafka.Teardown()
	a.Metadata.Teardown()
}

func (a *ApplicationImpl) Run() int {
	err := a.Create()
	if err != nil {
		return 10
	}
	defer a.Teardown()

	err = a.Server.Run()
	if err != nil {
		return 30
	}

	return 0
}

// not part of interface -- exposed for use by tests only

func (a *ApplicationImpl) ConstructConfigLoggingVaultTimestamp() error {
	a.Config, a.CustomConfig = config.New()
	a.Logging = logging.NewNoAcorn(a.Config)
	a.Vault = vault.NewNoAcorn(a.Config, a.Logging)
	if err := a.Config.Assemble(a.Logging); err != nil {
		return err
	}
	a.Logging.Setup()
	if err := vault.Execute(a.Vault); err != nil {
		return err
	}
	if err := a.Config.Setup(); err != nil {
		return err
	}
	a.Timestamp = timestamp.NewNoAcorn(time.Now)
	return nil
}

func (a *ApplicationImpl) ConstructRepositories() error {
	// construct the components that talk to the external world (must ensure correct order yourself), allowing for test overrides where needed
	if err := a.createGithub(); err != nil {
		return err
	}

	authProvider, err := authProvider.New(a.Config, a.CustomConfig, a.Logging, a.BaseRT)
	if err != nil {
		return err
	}
	a.AuthProvider = authProvider
	if err := a.AuthProvider.Setup(); err != nil {
		return err
	}

	if a.Metadata == nil {
		a.Metadata = metadata.New(a.Config, a.CustomConfig, a.Logging, a.Timestamp, a.AuthProvider)
	}
	if err := a.Metadata.Setup(); err != nil {
		return err
	}

	a.HostIP = hostip.New(a.Logging)
	if err := a.HostIP.Setup(); err != nil {
		return err
	}

	if a.Kafka == nil {
		a.Kafka = kafka.New(a.Config, a.CustomConfig, a.Logging, a.HostIP)
	}
	if err := a.Kafka.Setup(); err != nil {
		return err
	}

	if a.IdentityProvider == nil {
		a.IdentityProvider = idp.New(a.Config, a.CustomConfig, a.Logging)
	}
	if err := a.IdentityProvider.Setup(); err != nil {
		return err
	}

	a.Notifier = notifier.New(a.Config, a.CustomConfig, a.Logging)
	if err := a.Notifier.Setup(); err != nil {
		return err
	}

	a.Cache = cache.New(a.Config, a.CustomConfig, a.Logging, a.Timestamp)
	if err := a.Cache.Setup(); err != nil {
		return err
	}

	return nil
}

func (a *ApplicationImpl) ConstructServices() error {
	// construct the business logic components(must ensure correct order yourself)

	a.Mapper = mapper.New(a.Config, a.CustomConfig, a.Logging, a.Timestamp, a.Metadata)
	if err := a.Mapper.Setup(); err != nil {
		return err
	}

	a.Updater = updater.New(a.Config, a.CustomConfig, a.Logging, a.Timestamp, a.Kafka, a.Notifier, a.Mapper, a.Cache)
	if err := a.Updater.Setup(); err != nil {
		return err
	}
	a.Trigger = trigger.New(a.Config, a.CustomConfig, a.Logging, a.Timestamp, a.Updater)
	if err := a.Trigger.Setup(); err != nil {
		return err
	}

	a.Owners = owners.New(a.Config, a.Logging, a.Timestamp, a.Cache, a.Updater)
	if err := a.Owners.Setup(); err != nil {
		return err
	}

	a.Repositories = repositories.New(a.Config, a.CustomConfig, a.Logging, a.Timestamp, a.Cache, a.Updater, a.Owners)
	if err := a.Repositories.Setup(); err != nil {
		return err
	}

	a.Services = services.New(a.Config, a.CustomConfig, a.Logging, a.Timestamp, a.Cache, a.Updater, a.Repositories)
	if err := a.Services.Setup(); err != nil {
		return err
	}

	a.Validator = validator.New(a.Config, a.Repositories, a.Github, a.AuthProvider)

	if a.WebhooksHandler == nil {
		a.WebhooksHandler = webhookshandler.New(a.Config, a.Timestamp, a.Updater, a.Validator)
	}

	return nil
}

func (a *ApplicationImpl) ConstructControllers() error {
	// construct the components that handle incoming requests (must ensure correct order yourself)

	a.HealthCtl = healthctl.NewNoAcorn()
	a.SwaggerCtl = swaggerctl.NewNoAcorn()
	a.OwnerCtl = ownerctl.New(a.Config, a.CustomConfig, a.Logging, a.Timestamp, a.Owners)
	a.ServiceCtl = servicectl.New(a.Config, a.CustomConfig, a.Logging, a.Timestamp, a.Services)
	a.RepositoryCtl = repositoryctl.New(a.Config, a.CustomConfig, a.Logging, a.Timestamp, a.Repositories)
	a.WebhookCtl = webhookctl.New(a.Logging, a.Timestamp, a.WebhooksHandler)

	a.Server = server.New(a.Config, a.CustomConfig, a.Logging, a.IdentityProvider,
		a.HealthCtl, a.SwaggerCtl, a.OwnerCtl, a.ServiceCtl, a.RepositoryCtl, a.WebhookCtl)
	if err := a.Server.Setup(); err != nil {
		return err
	}

	return nil
}

func (a *ApplicationImpl) createGithub() error {
	if a.Github == nil {
		recorder := aurestrecorder.NewRecorderRoundTripper(http.DefaultTransport)
		a.BaseRT = recorder
		authTr, err := ghinstallation.New(recorder, a.CustomConfig.GithubAppId(), a.CustomConfig.GithubAppInstallationId(), a.CustomConfig.GithubAppJwtSigningKeyPEM())
		paginator := githubpagination.NewClient(authTr,
			githubpagination.WithPerPage(100),
			githubpagination.WithMaxNumOfPages(10),
		)
		if err != nil {
			return err
		}
		a.Github = githubclient.New(a.Timestamp, gogithub.NewClient(paginator))
	}
	return nil
}
