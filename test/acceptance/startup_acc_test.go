package acceptance

import (
	"github.com/Interhyp/go-backend-service-common/docs"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestStartup_ShouldBeHealthy(t *testing.T) {
	docs.Given("Given a valid configuration")

	docs.When("When the application is started")

	docs.Then("Then all components are present and of the correct type")
	require.NotNil(t, application.Config)
	require.NotNil(t, application.CustomConfig)
	require.NotNil(t, application.Logging)
	require.NotNil(t, application.Vault)
	require.NotNil(t, application.Kafka)
	require.NotNil(t, application.Metadata)
	require.NotNil(t, application.HostIP)

	require.NotNil(t, application.Cache)
	require.NotNil(t, application.Mapper)
	require.NotNil(t, application.Trigger)
	require.NotNil(t, application.Updater)
	require.NotNil(t, application.Owners)
	require.NotNil(t, application.Services)
	require.NotNil(t, application.Repositories)

	require.NotNil(t, application.HealthCtl)
	require.NotNil(t, application.SwaggerCtl)
	require.NotNil(t, application.OwnerCtl)
	require.NotNil(t, application.ServiceCtl)
	require.NotNil(t, application.RepositoryCtl)
	require.NotNil(t, application.WebhookCtl)

	require.NotNil(t, application.Server)

	docs.Then("And the application reports as healthy")
	response, err := tstPerformGet("/management/health", tstUnauthenticated())
	tstAssert(t, response, err, http.StatusOK, "health.json")

	response, err = tstPerformGet("/health", tstUnauthenticated())
	tstAssert(t, response, err, http.StatusOK, "health.json")

	response, err = tstPerformGet("/", tstUnauthenticated())
	tstAssert(t, response, err, http.StatusOK, "health.json")
}
