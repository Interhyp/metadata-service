package acceptance

import (
	"context"
	application2 "github.com/Interhyp/metadata-service/acorns/application"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/Interhyp/metadata-service/acorns/service"
	"github.com/Interhyp/metadata-service/internal/repository/config"
	"github.com/Interhyp/metadata-service/internal/service/mapper"
	"github.com/Interhyp/metadata-service/internal/service/trigger"
	"github.com/Interhyp/metadata-service/internal/service/updater"
	"github.com/Interhyp/metadata-service/internal/web/app"
	"github.com/Interhyp/metadata-service/internal/web/server"
	"github.com/Interhyp/metadata-service/test/acceptance/bitbucketmock"
	"github.com/Interhyp/metadata-service/test/acceptance/idpmock"
	"github.com/Interhyp/metadata-service/test/acceptance/kafkamock"
	"github.com/Interhyp/metadata-service/test/acceptance/metadatamock"
	"github.com/Interhyp/metadata-service/test/acceptance/vaultmock"
	auacorn "github.com/StephanHCB/go-autumn-acorn-registry"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	libconfig "github.com/StephanHCB/go-backend-service-common/repository/config"
	"github.com/StephanHCB/go-backend-service-common/repository/logging"
	"github.com/StephanHCB/go-backend-service-common/web/middleware/security"
	"github.com/rs/zerolog/log"
	"net/http/httptest"
	"time"
)

// placing these here because they are package global

var (
	ts *httptest.Server

	configImpl   *libconfig.ConfigImpl
	loggingImpl  *logging.LoggingImpl
	vaultImpl    *vaultmock.VaultImpl
	metadataImpl *metadatamock.Impl
	kafkaImpl    *kafkamock.Impl
	idpImpl      *idpmock.Impl
	bbImpl       *bitbucketmock.BitbucketMock

	application application2.Application
	appCtx      context.Context
)

const validConfigurationPath = "../resources/valid-config.yaml"

func fakeNow() time.Time {
	return time.Date(2022, 11, 6, 18, 14, 10, 0, time.UTC)
}

// use a special configuration and wire in mocks for most repositories
func tstSetup(configPath string) error {
	application = app.New()
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

	application.Register()

	application.Create()
	// can now manipulate the registry by inserting custom instances
	registry := auacorn.Registry.(*auacorn.AcornRegistryImpl)
	registry.CreateOverride(librepo.ConfigurationAcornName, configImpl)
	registry.CreateOverride(librepo.LoggingAcornName, loggingImpl)
	registry.CreateOverride(librepo.VaultAcornName, vaultImpl)
	registry.CreateOverride(repository.MetadataAcornName, metadataImpl)
	registry.CreateOverride(repository.KafkaAcornName, kafkaImpl)
	registry.CreateOverride(repository.IdentityProviderAcornName, idpImpl)
	registry.CreateOverride(repository.BitbucketAcornName, bbImpl)

	registry.SkipAssemble(loggingImpl) // already assembled
	registry.SkipAssemble(configImpl)  // would attempt to read config
	err = application.Assemble()
	if err != nil {
		return err
	}

	// other features that need switching off or changing

	timestamp := registry.GetAcornByName(librepo.TimestampAcornName).(librepo.Timestamp)
	timestamp.MockResponse(fakeNow)

	triggerImpl := registry.GetAcornByName(service.TriggerAcornName).(*trigger.Impl)
	triggerImpl.SkipStart = true // do not start cron job
	triggerImpl.Now = fakeNow

	mapperImpl := registry.GetAcornByName(service.MapperAcornName).(*mapper.Impl)
	mapperImpl.Now = fakeNow

	updaterImpl := registry.GetAcornByName(service.UpdaterAcornName).(*updater.Impl)
	updaterImpl.Now = fakeNow

	metadataImpl.Now = fakeNow

	security.Now = fakeNow

	registry.SkipSetup(loggingImpl)
	registry.SkipSetup(configImpl)
	registry.Setup()

	tstSetupHttpTestServer()
	return nil
}

func tstSetupConfig(configPath string) error {
	configImpl = config.New().(*libconfig.ConfigImpl)
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
	application.(*app.ApplicationImpl).Server.WireUp(appCtx)
	ts = httptest.NewServer(application.(*app.ApplicationImpl).Server.(*server.Impl).Router)
}

func tstShutdown() {
	ts.Close()
}

func tstReset() {
	metadataImpl.Reset()
	kafkaImpl.Reset()
}
