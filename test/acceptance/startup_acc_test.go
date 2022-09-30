package acceptance

import (
	"github.com/Interhyp/metadata-service/acorns/controller"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/Interhyp/metadata-service/acorns/service"
	"github.com/Interhyp/metadata-service/docs"
	"github.com/Interhyp/metadata-service/web/app"
	auacorn "github.com/StephanHCB/go-autumn-acorn-registry"
	libcontroller "github.com/StephanHCB/go-backend-service-common/acorns/controller"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestStartup_ShouldBeHealthy(t *testing.T) {
	docs.Given("Given a valid configuration")

	docs.When("When the application is started")

	docs.Then("Then all components are present and of the correct type")
	appImpl := application.(*app.ApplicationImpl)
	registry := auacorn.Registry.(*auacorn.AcornRegistryImpl)
	require.NotNil(t, registry.GetAcornByName(librepo.ConfigurationAcornName).(librepo.Configuration))
	require.NotNil(t, registry.GetAcornByName(librepo.LoggingAcornName).(librepo.Logging))
	require.NotNil(t, registry.GetAcornByName(repository.VaultAcornName).(repository.Vault))
	require.NotNil(t, registry.GetAcornByName(repository.KafkaAcornName).(repository.Kafka))
	require.NotNil(t, registry.GetAcornByName(repository.MetadataAcornName).(repository.Metadata))
	require.NotNil(t, registry.GetAcornByName(repository.HostIPAcornName).(repository.HostIP))

	require.NotNil(t, registry.GetAcornByName(service.CacheAcornName).(service.Cache))
	require.NotNil(t, registry.GetAcornByName(service.MapperAcornName).(service.Mapper))
	require.NotNil(t, registry.GetAcornByName(service.TriggerAcornName).(service.Trigger))
	require.NotNil(t, registry.GetAcornByName(service.UpdaterAcornName).(service.Updater))
	require.NotNil(t, registry.GetAcornByName(service.OwnersAcornName).(service.Owners))
	require.NotNil(t, registry.GetAcornByName(service.ServicesAcornName).(service.Services))
	require.NotNil(t, registry.GetAcornByName(service.RepositoriesAcornName).(service.Repositories))

	require.NotNil(t, registry.GetAcornByName(libcontroller.HealthControllerAcornName).(libcontroller.HealthController))
	require.NotNil(t, registry.GetAcornByName(libcontroller.SwaggerControllerAcornName).(libcontroller.SwaggerController))
	require.NotNil(t, registry.GetAcornByName(controller.OwnerControllerAcornName).(controller.OwnerController))
	require.NotNil(t, registry.GetAcornByName(controller.ServiceControllerAcornName).(controller.ServiceController))
	require.NotNil(t, registry.GetAcornByName(controller.RepositoryControllerAcornName).(controller.RepositoryController))
	require.NotNil(t, registry.GetAcornByName(controller.WebhookControllerAcornName).(controller.WebhookController))

	require.NotNil(t, appImpl.Server)

	docs.Then("And the application reports as healthy")
	response, err := tstPerformGet("/management/health", tstUnauthenticated())
	tstAssert(t, response, err, http.StatusOK, "health.json")

	response, err = tstPerformGet("/health", tstUnauthenticated())
	tstAssert(t, response, err, http.StatusOK, "health.json")

	response, err = tstPerformGet("/", tstUnauthenticated())
	tstAssert(t, response, err, http.StatusOK, "health.json")
}
