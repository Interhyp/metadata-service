package acceptance

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/repository/config"
	"github.com/Interhyp/metadata-service/internal/repository/notifier"
	"github.com/Interhyp/metadata-service/internal/service/trigger"
	"github.com/Interhyp/metadata-service/internal/web/app"
	"github.com/Interhyp/metadata-service/internal/web/controller/webhookctl"
	"github.com/Interhyp/metadata-service/internal/web/server"
	"github.com/Interhyp/metadata-service/test/mock/bitbucketmock"
	"github.com/Interhyp/metadata-service/test/mock/idpmock"
	"github.com/Interhyp/metadata-service/test/mock/kafkamock"
	"github.com/Interhyp/metadata-service/test/mock/metadatamock"
	"github.com/Interhyp/metadata-service/test/mock/notifiermock"
	"github.com/Interhyp/metadata-service/test/mock/vaultmock"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
	libconfig "github.com/StephanHCB/go-backend-service-common/repository/config"
	"github.com/StephanHCB/go-backend-service-common/repository/logging"
	"github.com/StephanHCB/go-backend-service-common/repository/timestamp"
	"github.com/StephanHCB/go-backend-service-common/web/middleware/security"
	"github.com/rs/zerolog/log"
	"net/http/httptest"
	"time"
)

// placing these here because they are package global

type ApplicationWithMocksImpl struct {
	app.ApplicationImpl
}

var (
	ts *httptest.Server

	configImpl       *libconfig.ConfigImpl
	customConfigImpl *config.CustomConfigImpl
	loggingImpl      *logging.LoggingImpl
	vaultImpl        *vaultmock.VaultImpl
	metadataImpl     *metadatamock.Impl
	kafkaImpl        *kafkamock.Impl
	idpImpl          *idpmock.Impl
	bbImpl           *bitbucketmock.BitbucketMock
	notifierImpl     *notifier.Impl

	application *ApplicationWithMocksImpl
	appCtx      context.Context
)

const validConfigurationPath = "../resources/valid-config.yaml"

func fakeNow() time.Time {
	return time.Date(2022, 11, 6, 18, 14, 10, 0, time.UTC)
}

func (a *ApplicationWithMocksImpl) Create() error {
	a.ConstructConfigLoggingVaultTimestamp_ForTesting()

	// prefill mocks as overrides
	a.Metadata = metadataImpl
	a.Kafka = kafkaImpl
	a.IdentityProvider = idpImpl
	a.Bitbucket = bbImpl

	// now can use normal construct functions, they respect the prefilled mocks
	if err := a.ConstructRepositories(); err != nil {
		return err
	}
	if err := a.ConstructServices(); err != nil {
		return err
	}
	if err := a.ConstructControllers(); err != nil {
		return err
	}

	a.WebhookCtl.(*webhookctl.Impl).EnableAsync = false

	return nil
}

func (a *ApplicationWithMocksImpl) ConstructConfigLoggingVaultTimestamp_ForTesting() {
	// construct and set up config, logging, vault, timestamp
	a.Config = configImpl
	a.CustomConfig = customConfigImpl
	a.Logging = loggingImpl
	a.Vault = vaultImpl
	a.Timestamp = timestamp.NewNoAcorn(fakeNow)
}

func (a *ApplicationWithMocksImpl) Teardown() {
	// reverse order (must ensure correct order yourself, but most components will not have a teardown method)
	a.Trigger.Teardown()
	a.Kafka.Teardown()
	a.Metadata.Teardown()
}

// use a special configuration and wire in mocks for most repositories
func tstSetup(configPath string) error {
	application = &ApplicationWithMocksImpl{}
	err := tstSetupConfig(configPath)
	if err != nil {
		return err
	}
	tstSetupLogging()

	vaultImpl = vaultmock.New().(*vaultmock.VaultImpl)
	metadataImpl = metadatamock.New().(*metadatamock.Impl)
	kafkaImpl = kafkamock.New().(*kafkamock.Impl)
	idpImpl = idpmock.New().(*idpmock.Impl)
	bbImpl = bitbucketmock.New().(*bitbucketmock.BitbucketMock)

	metadataImpl.Now = fakeNow

	err = application.Create()
	if err != nil {
		return err
	}

	application.Trigger.(*trigger.Impl).SkipStart = true // do not start cron job

	notifierImpl = application.Notifier.(*notifier.Impl)
	notifierImpl.SkipAsync = true

	security.Now = fakeNow

	for identifier, _ := range notifierImpl.Clients {
		notifierImpl.Clients[identifier] = &notifiermock.NotifierClientMock{SentNotifications: make([]string, 0)}
	}

	tstSetupHttpTestServer()
	return nil
}

func tstSetupConfig(configPath string) error {
	impl, cImpl := config.New()
	configImpl = impl.(*libconfig.ConfigImpl)
	customConfigImpl = cImpl.(*config.CustomConfigImpl)
	auconfigenv.LocalConfigFileName = configPath
	err := configImpl.Read()
	if err != nil {
		return err
	}
	// do not wish to validate configuration, so setting parsed values directly
	configImpl.ObtainPredefinedValues()
	configImpl.CustomConfiguration.Obtain(auconfigenv.Get)

	customConfigImpl := configImpl.CustomConfiguration.(*config.CustomConfigImpl)
	// and can override configuration values here
	customConfigImpl.VUpdateJobTimeoutSeconds = 1
	return nil
}

func tstSetupLogging() {
	loggingImpl = logging.New().(*logging.LoggingImpl)
	loggingImpl.SetupForTesting()
	appCtx = log.Logger.WithContext(context.Background())
	configImpl.Logging = loggingImpl
	loggingImpl.Configuration = configImpl
}

func tstSetupHttpTestServer() {
	application.Server.WireUp(appCtx)
	ts = httptest.NewServer(application.Server.(*server.Impl).Router)
}

func tstShutdown() {
	ts.Close()
}

func tstReset() {
	metadataImpl.Reset()
	kafkaImpl.Reset()
	for _, client := range notifierImpl.Clients {
		client.(*notifiermock.NotifierClientMock).Reset()
	}
	bbImpl.Recording = nil
	bbImpl.PRHead = ""
	bbImpl.ChangedFilesResponse = nil
}
