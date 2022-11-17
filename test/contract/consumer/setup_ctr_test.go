package consumer

import (
	"github.com/Interhyp/metadata-service/internal/repository/config"
	"github.com/Interhyp/metadata-service/web/app"
	auacorn "github.com/StephanHCB/go-autumn-acorn-registry"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
	libconfig "github.com/StephanHCB/go-backend-service-common/repository/config"
	"github.com/StephanHCB/go-backend-service-common/repository/logging"
)

const contractTestConfigurationPath = "../../resources/contract-test-config.yaml"

func tstSetup() error {
	application := app.New().(*app.ApplicationImpl)

	// setup test configuration
	configImpl := config.New().(*libconfig.ConfigImpl)
	auconfigenv.LocalConfigFileName = contractTestConfigurationPath
	err := configImpl.Read()
	if err != nil {
		return err
	}
	// intentionally not validating the configuration
	configImpl.ObtainPredefinedValues()
	configImpl.CustomConfiguration.Obtain(auconfigenv.Get)

	// setup logging
	loggingImpl := logging.New().(*logging.LoggingImpl)
	loggingImpl.Configuration = configImpl
	loggingImpl.Setup()
	configImpl.Logging = loggingImpl

	application.Register()
	application.Create()
	// config has to be read again as application.Create() reinitialises without reading
	err = configImpl.Read()
	if err != nil {
		return err
	}
	// now can manipulate the registry by inserting custom instances
	registry := auacorn.Registry.(*auacorn.AcornRegistryImpl)
	registry.CreateOverride("configuration", configImpl)
	registry.CreateOverride("logging", loggingImpl)

	registry.SkipAssemble(loggingImpl) // already assembled
	registry.SkipAssemble(configImpl)  // would attempt to read config
	err = application.Assemble()

	return err
}
