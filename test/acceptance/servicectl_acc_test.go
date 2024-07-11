package acceptance

import (
	"encoding/json"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/types"
	"net/http"
	"strings"
	"testing"

	"github.com/StephanHCB/go-backend-service-common/docs"
	"github.com/stretchr/testify/require"
)

// get services

func TestGETServices_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request the list of services")
	response, err := tstPerformGet("/rest/api/v1/services", token)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "services.json")
}

// get service

func TestGETService_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request a single existing service")
	response, err := tstPerformGet("/rest/api/v1/services/some-service-backend", token)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "service.json")
}

func TestGETService_NotFound(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request a single service that does not exist")
	response, err := tstPerformGet("/rest/api/v1/services/unicorn", token)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusNotFound, "service-notfound.json")
}

// create service

func TestPOSTService_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a valid service that does not exist")
	body := tstService("whatever")
	response, err := tstPerformPost("/rest/api/v1/services/whatever", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusCreated, "service-create.json")

	docs.Then("And the service has been correctly written, committed and pushed")
	filename := "owners/some-owner/services/whatever.yaml"
	require.Equal(t, tstServiceExpectedYaml("whatever"), metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the service has been cached and can be read again")
	readAgain, err := tstPerformGet("/rest/api/v1/services/whatever", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "service-create.json")

	docs.Then("And a kafka message notifying other instances of the creation has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstServiceExpectedKafka("whatever"), string(actual))

	docs.Then("And a notification has been sent to all matching owners")
	payload := tstNewServicePayload("whatever")
	hasSentNotification(t, "receivesCreate", "whatever", types.CreatedEvent, types.ServicePayload, &payload)
	hasSentNotification(t, "receivesService", "whatever", types.CreatedEvent, types.ServicePayload, &payload)
}

func TestPOSTService_InvalidName(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a service with an invalid name")

	body := tstService("CapitalsAreForbidden")
	response, err := tstPerformPost("/rest/api/v1/services/CapitalsAreForbidden", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-invalid-name.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTService_ProhibitedName(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a service with an invalid name")

	body := tstService("some-service")
	response, err := tstPerformPost("/rest/api/v1/services/some-service", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-invalid-name.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTService_InvalidBody(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a service with an invalid body")
	body := []byte(`{"invalid json syntax":{,{`)
	response, err := tstPerformRawWithBody(http.MethodPost, "/rest/api/v1/services/post-service-invalid-syntax", token, body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-invalid-syntax.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTService_InvalidValues(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a service with invalid values in the body")
	body := tstService("post-service-invalid-values")
	body.Repositories = []string{
		"whatever.helm-deployment", // cross ref to other service not allowed
	}
	response, err := tstPerformPost("/rest/api/v1/services/post-service-invalid-values", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-create-invalid-values.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTService_NonexistentRepository(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a service referring to a repository that does not exist")
	body := tstService("post-service-invalid-repo")
	response, err := tstPerformPost("/rest/api/v1/services/post-service-invalid-repo", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-create-repo-missing.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTService_NonexistentOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a service referring to an owner that does not exist")
	body := tstService("whatever")
	body.Owner = "not-there"
	response, err := tstPerformPost("/rest/api/v1/services/whatever", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-create-owner-missing.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTService_Unauthenticated(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user (no token)")
	token := tstUnauthenticated()

	docs.When("When they request the creation of a valid service that does not exist")
	body := tstService("whatever")
	response, err := tstPerformPost("/rest/api/v1/services/whatever", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTService_ExpiredToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with an expired token")
	token := tstExpiredAdminToken()

	docs.When("When they request the creation of a valid service that does not exist")
	body := tstService("whatever")
	response, err := tstPerformPost("/rest/api/v1/services/whatever", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTService_NonAdminToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with a valid token without the admin role")
	token := tstValidUserToken()

	docs.When("When they request the creation of a valid service that does not exist")
	body := tstService("whatever")
	response, err := tstPerformPost("/rest/api/v1/services/whatever", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusForbidden, "forbidden.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTService_Duplicate(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a valid service that already exists")
	body := tstService("some-service-backend")
	response, err := tstPerformPost("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusConflict, "service-create-duplicate.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTService_GitServerDown(t *testing.T) {
	tstReset()

	docs.Given("Given the git server is down")
	metadataImpl.SimulateRemoteFailure = true

	docs.Given("And given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a valid service")
	body := tstService("whatever")
	response, err := tstPerformPost("/rest/api/v1/services/whatever", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadGateway, "bad-gateway.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTService_GitHookDeclined(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a valid service, but supply an invalid issue")
	body := tstService("whatever")
	body.JiraIssue = "INVALID-12345"
	response, err := tstPerformPost("/rest/api/v1/services/whatever", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "receive-hook-declined.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTService_ImplementationCrossrefAllowed(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.Given("Given an existing repositories crossref.helm-deployment and not-crossref.implementation")
	deplRepoBody := tstRepository()
	deplRepoResponse, err := tstPerformPost("/rest/api/v1/repositories/crossref.helm-deployment", token, &deplRepoBody)
	require.Nil(t, err)
	require.Equal(t, http.StatusCreated, deplRepoResponse.status)

	implRepoBody := tstRepository()
	implRepoResponse, err := tstPerformPost("/rest/api/v1/repositories/not-crossref.implementation", token, &implRepoBody)
	require.Nil(t, err)
	require.Equal(t, http.StatusCreated, implRepoResponse.status)

	docs.When("When they request the creation of a new service that references a pre-existing implementation repository with a different primary name")
	body := tstService("crossref")
	body.Repositories = []string{
		"crossref.helm-deployment",
		"not-crossref.implementation",
	}
	response, err := tstPerformPost("/rest/api/v1/services/crossref", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusCreated, "service-create-crossref.json")

	docs.Then("And the service has been correctly written, committed and pushed")
	filename := "owners/some-owner/services/crossref.yaml"
	expectedYaml := strings.ReplaceAll(tstServiceExpectedYaml("crossref"), "crossref/implementation", "not-crossref/implementation")
	require.Equal(t, expectedYaml, metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)
}

// update full service

func TestPUTService_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid update of an existing service")
	body := tstService("some-service-backend")
	response, err := tstPerformPut("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "service-update.json")

	docs.Then("And the service has been correctly written, committed and pushed")
	filename := "owners/some-owner/services/some-service-backend.yaml"
	require.Equal(t, tstServiceExpectedYaml("some-service-backend"), metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the service has been cached and can be read again")
	readAgain, err := tstPerformGet("/rest/api/v1/services/some-service-backend", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "service-update.json")

	docs.Then("And a kafka message notifying other instances of the update has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstServiceExpectedKafka("some-service-backend"), string(actual))

	docs.Then("And a notification has been sent to all matching owners")
	payload := tstNewServicePayload("some-service-backend")
	hasSentNotification(t, "receivesModified", "some-service-backend", types.ModifiedEvent, types.ServicePayload, &payload)
	hasSentNotification(t, "receivesService", "some-service-backend", types.ModifiedEvent, types.ServicePayload, &payload)
}

func TestPUTService_NoChangeSuccess(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid update of an existing service that does not make actual changes")
	metadataImpl.SimulateUnchangedFailure = true
	body := tstServiceUnchanged("some-service-backend")
	response, err := tstPerformPut("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "service-unchanged.json")

	docs.Then("And no commit has been made and pushed (because there were no changes)")
	filename := "owners/some-owner/services/some-service-backend.yaml"
	require.Equal(t, tstServiceUnchangedExpectedYaml("some-service-backend"), metadataImpl.ReadContents(filename))
	require.False(t, metadataImpl.FilesCommitted[filename])
	require.False(t, metadataImpl.Pushed)

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestPUTService_DoesNotExist(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt a valid update of a service that does not exist")
	body := tstService("does-not-exist")
	response, err := tstPerformPut("/rest/api/v1/services/does-not-exist", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusNotFound, "service-notfound-doesnotexist.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestPUTService_NonexistentRepository(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a service referring to a repository that does not exist")
	body := tstService("some-service-backend")
	body.Repositories = []string{"some-service-backend.api"}
	response, err := tstPerformPut("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-update-repo-missing.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTService_NonexistentOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a service referring to an owner that does not exist")
	body := tstService("some-service-backend")
	body.Owner = "not-there"
	response, err := tstPerformPut("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-update-owner-missing.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTService_InvalidBody(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a service with a syntactically invalid body")
	body := []byte(`{"invalid json syntax":{,{`)
	response, err := tstPerformRawWithBody(http.MethodPut, "/rest/api/v1/services/some-service-backend", token, body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-invalid-syntax.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTService_InvalidValues(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a service with invalid values in the body")
	body := tstService("some-service-backend")
	body.Owner = ""
	body.AlertTarget = ""
	body.TimeStamp = ""
	response, err := tstPerformPut("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-update-invalid-values.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTService_Unauthenticated(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user (no token)")
	token := tstUnauthenticated()

	docs.When("When they request an update of a service")
	body := tstService("some-service-backend")
	response, err := tstPerformPut("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTService_ExpiredToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with an expired token")
	token := tstExpiredAdminToken()

	docs.When("When they request an update of a service")
	body := tstService("some-service-backend")
	response, err := tstPerformPut("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTService_NonAdminToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with a valid token without the admin role")
	token := tstValidUserToken()

	docs.When("When they request an update of a service")
	body := tstService("some-service-backend")
	response, err := tstPerformPut("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusForbidden, "forbidden.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTService_Conflict(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a service based on an outdated commit hash")
	body := tstService("some-service-backend")
	body.CommitHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	body.TimeStamp = "2019-01-01T00:00:00Z"
	response, err := tstPerformPut("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusConflict, "service-update-conflict.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTService_GitServerDown(t *testing.T) {
	tstReset()

	docs.Given("Given the git server is down")
	metadataImpl.SimulateRemoteFailure = true

	docs.Given("And given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a service")
	body := tstService("some-service-backend")
	response, err := tstPerformPut("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadGateway, "bad-gateway.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTService_GitHookDeclined(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a service, but supply an invalid issue")
	body := tstService("some-service-backend")
	body.JiraIssue = "INVALID-12345"
	response, err := tstPerformPut("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "receive-hook-declined.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTService_ChangeOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid update of an existing service that changes its owner")
	body := tstService("some-service-backend")
	body.Owner = "deleteme"
	response, err := tstPerformPut("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "service-update-newowner.json")

	docs.Then("And the service with its repositories has been correctly moved, committed and pushed")
	filename1old := "owners/some-owner/services/some-service-backend.yaml"
	filename1 := "owners/deleteme/services/some-service-backend.yaml"
	filename2old := "owners/some-owner/repositories/some-service-backend.helm-deployment.yaml"
	filename2 := "owners/deleteme/repositories/some-service-backend.helm-deployment.yaml"
	filename3old := "owners/some-owner/repositories/some-service-backend.helm-deployment.yaml"
	filename3 := "owners/deleteme/repositories/some-service-backend.helm-deployment.yaml"
	require.Equal(t, tstServiceExpectedYaml("some-service-backend"), metadataImpl.ReadContents(filename1))
	require.True(t, metadataImpl.FilesCommitted[filename1])
	require.True(t, metadataImpl.FilesCommitted[filename1old])
	require.True(t, metadataImpl.FilesCommitted[filename2])
	require.True(t, metadataImpl.FilesCommitted[filename2old])
	require.True(t, metadataImpl.FilesCommitted[filename3])
	require.True(t, metadataImpl.FilesCommitted[filename3old])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the service has been cached and can be read again, returning the correct owner")
	readAgain, err := tstPerformGet("/rest/api/v1/services/some-service-backend", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "service-update-newowner.json")

	docs.Then("And a kafka message notifying other instances of the update has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstServiceMovedExpectedKafka("some-service-backend"), string(actual))
}

func TestPUTService_ChangeOwner_GitHookDeclined(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a service that changes its owner, but supply an invalid issue")
	body := tstService("some-service-backend")
	body.Owner = "deleteme"
	body.JiraIssue = "INVALID-12345"
	response, err := tstPerformPut("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "receive-hook-declined.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And when the service is requested again, it still has the original owner")
	readAgain, err := tstPerformGet("/rest/api/v1/services/some-service-backend", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "service-original.json")
}

func TestPUTService_ImplementationCrossrefAllowed(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.Given("Given existing repositories crossref.helm-deployment and not-crossref.implementation")
	deplRepoBody := tstRepository()
	deplRepoResponse, err := tstPerformPost("/rest/api/v1/repositories/crossref.helm-deployment", token, &deplRepoBody)
	require.Nil(t, err)
	require.Equal(t, http.StatusCreated, deplRepoResponse.status)

	implRepoBody := tstRepository()
	implRepoResponse, err := tstPerformPost("/rest/api/v1/repositories/not-crossref.implementation", token, &implRepoBody)
	require.Nil(t, err)
	require.Equal(t, http.StatusCreated, implRepoResponse.status)

	docs.Given("Given an existing service that references only the helm-deployment repository")
	createBody := tstService("crossref")
	createBody.Repositories = []string{
		"crossref.helm-deployment",
	}
	createResponse, err := tstPerformPost("/rest/api/v1/services/crossref", token, &createBody)
	require.Nil(t, err)
	require.Equal(t, http.StatusCreated, createResponse.status)
	// need commit hash from response, so we can perform valid update
	var tmp openapi.ServiceDto
	err = json.Unmarshal([]byte(createResponse.body), &tmp)
	require.Nil(t, err)

	docs.When("When they perform an update of the service, adding a cross-referenced implementation repository")
	body := tstService("crossref")
	body.Repositories = []string{
		"crossref.helm-deployment",
		"not-crossref.implementation",
	}
	body.CommitHash = tmp.CommitHash
	response, err := tstPerformPut("/rest/api/v1/services/crossref", token, &body)

	docs.Then("Then the request is successful")
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, response.status)

	docs.Then("And the service has been correctly written, committed and pushed")
	filename := "owners/some-owner/services/crossref.yaml"
	expectedYaml := strings.ReplaceAll(tstServiceExpectedYaml("crossref"), "crossref/implementation", "not-crossref/implementation")
	require.Equal(t, expectedYaml, metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)
}

// patch service

func TestPATCHService_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid patch of an existing service")
	body := tstServicePatch()
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "service-patch.json")

	docs.Then("And the service has been correctly written, committed and pushed")
	filename := "owners/some-owner/services/some-service-backend.yaml"
	require.Equal(t, tstServiceExpectedYaml("some-service-backend"), metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the service has been cached and can be read again")
	readAgain, err := tstPerformGet("/rest/api/v1/services/some-service-backend", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "service-patch.json")

	docs.Then("And a kafka message notifying other instances of the update has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstServiceExpectedKafka("some-service-backend"), string(actual))

	docs.Then("And a notification has been sent to all matching owners")
	payload := tstUpdatedServicePayload("some-service-backend")
	hasSentNotification(t, "receivesModified", "some-service-backend", types.ModifiedEvent, types.ServicePayload, &payload)
	hasSentNotification(t, "receivesService", "some-service-backend", types.ModifiedEvent, types.ServicePayload, &payload)
}

func TestPATCHService_NoChangeSuccess(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid patch of an existing service that does not make actual changes")
	metadataImpl.SimulateUnchangedFailure = true
	body := tstServiceUnchangedPatch()
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "service-unchanged.json")

	docs.Then("And no commit has been made and pushed (because there were no changes)")
	filename := "owners/some-owner/services/some-service-backend.yaml"
	require.Equal(t, tstServiceUnchangedExpectedYaml("some-service-backend"), metadataImpl.ReadContents(filename))
	require.False(t, metadataImpl.FilesCommitted[filename])
	require.False(t, metadataImpl.Pushed)

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestPATCHService_DoesNotExist(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to patch a service that does not exist")
	body := tstServicePatch()
	response, err := tstPerformPatch("/rest/api/v1/services/does-not-exist", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusNotFound, "service-notfound-doesnotexist.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestPATCHService_NonexistentRepository(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to patch a service referring to a repository that does not exist")
	body := tstServicePatch()
	body.Repositories = []string{"some-service-backend.api"}
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-update-repo-missing.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHService_NonexistentOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to patch a service referring to an owner that does not exist")
	body := tstServicePatch()
	body.Owner = ptr("not-there")
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-update-owner-missing.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHService_InvalidBody(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request a patch of a service with a syntactically invalid body")
	body := []byte(`},},},},},},},}`)
	response, err := tstPerformRawWithBody(http.MethodPatch, "/rest/api/v1/services/some-service-backend", token, body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-invalid-syntax.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHService_InvalidValues(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request a patch of a service with invalid values in the body")
	body := tstServicePatch()
	body.Owner = ptr("")       // invalid
	body.AlertTarget = ptr("") // invalid (nil would be valid)
	body.CommitHash = ""
	body.TimeStamp = ""
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-patch-invalid-values.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHService_Unauthenticated(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user (no token)")
	token := tstUnauthenticated()

	docs.When("When they attempt to patch a service")
	body := tstServicePatch()
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHService_ExpiredToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with an expired token")
	token := tstExpiredAdminToken()

	docs.When("When they attempt to patch a service")
	body := tstServicePatch()
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHService_NonAdminToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with a valid token without the admin role")
	token := tstValidUserToken()

	docs.When("When they attempt to patch a service")
	body := tstServicePatch()
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusForbidden, "forbidden.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHService_Conflict(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to patch a service based on an outdated commit hash")
	body := tstServicePatch()
	body.CommitHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	body.TimeStamp = "2019-01-01T00:00:00Z"
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusConflict, "service-patch-conflict.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHService_GitServerDown(t *testing.T) {
	tstReset()

	docs.Given("Given the git server is down")
	metadataImpl.SimulateRemoteFailure = true

	docs.Given("And given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to patch a service")
	body := tstServicePatch()
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadGateway, "bad-gateway.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHService_GitHookDeclined(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a service, but supply an invalid issue")
	body := tstServicePatch()
	body.JiraIssue = "INVALID-12345"
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "receive-hook-declined.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHService_ChangeOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid patch of an existing service that changes its owner")
	body := tstServicePatch()
	body.Owner = ptr("deleteme")
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "service-patch-newowner.json")

	docs.Then("And the service with its repositories has been correctly moved, committed and pushed")
	filename1old := "owners/some-owner/services/some-service-backend.yaml"
	filename1 := "owners/deleteme/services/some-service-backend.yaml"
	filename2old := "owners/some-owner/repositories/some-service-backend.helm-deployment.yaml"
	filename2 := "owners/deleteme/repositories/some-service-backend.helm-deployment.yaml"
	filename3old := "owners/some-owner/repositories/some-service-backend.implementation.yaml"
	filename3 := "owners/deleteme/repositories/some-service-backend.implementation.yaml"
	require.Equal(t, tstServiceExpectedYaml("some-service-backend"), metadataImpl.ReadContents(filename1))
	require.True(t, metadataImpl.FilesCommitted[filename1])
	require.True(t, metadataImpl.FilesCommitted[filename1old])
	require.True(t, metadataImpl.FilesCommitted[filename2])
	require.True(t, metadataImpl.FilesCommitted[filename2old])
	require.True(t, metadataImpl.FilesCommitted[filename3])
	require.True(t, metadataImpl.FilesCommitted[filename3old])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the service has been cached and can be read again, returning the correct owner")
	readAgain, err := tstPerformGet("/rest/api/v1/services/some-service-backend", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "service-patch-newowner.json")

	docs.Then("And a kafka message notifying other instances of the update has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstServiceMovedExpectedKafka("some-service-backend"), string(actual))
}

func TestPATCHService_ChangeSpec(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid patch of an existing service that changes its spec")
	body := tstServicePatch()
	body.Spec = &openapi.ServiceSpecDto{
		DependsOn:    []string{"some-service", "other-service"},
		ProvidesApis: []string{},
		ConsumesApis: []string{"some-api"},
	}
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "service-patch-spec.json")

	docs.Then("And the service has been cached and can be read again, returning the correct spec")
	readAgain, err := tstPerformGet("/rest/api/v1/services/some-service-backend", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "service-patch-spec.json")
}

func TestPATCHService_ImplementationCrossrefAllowed(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.Given("Given existing repositories crossref.helm-deployment and not-crossref.implementation")
	deplRepoBody := tstRepository()
	deplRepoResponse, err := tstPerformPost("/rest/api/v1/repositories/crossref.helm-deployment", token, &deplRepoBody)
	require.Nil(t, err)
	require.Equal(t, http.StatusCreated, deplRepoResponse.status)

	implRepoBody := tstRepository()
	implRepoResponse, err := tstPerformPost("/rest/api/v1/repositories/not-crossref.implementation", token, &implRepoBody)
	require.Nil(t, err)
	require.Equal(t, http.StatusCreated, implRepoResponse.status)

	docs.Given("Given an existing service that references only the helm-deployment repository")
	createBody := tstService("crossref")
	createBody.Repositories = []string{
		"crossref.helm-deployment",
	}
	createResponse, err := tstPerformPost("/rest/api/v1/services/crossref", token, &createBody)
	require.Nil(t, err)
	require.Equal(t, http.StatusCreated, createResponse.status)
	// need commit hash from response, so we can perform valid update
	var tmp openapi.ServiceDto
	err = json.Unmarshal([]byte(createResponse.body), &tmp)
	require.Nil(t, err)

	docs.When("When they perform a valid patch of the service that adds a crossreferenced repository")
	body := tstServicePatch()
	body.Repositories = []string{
		"crossref.helm-deployment",
		"not-crossref.implementation",
	}
	body.CommitHash = tmp.CommitHash
	response, err := tstPerformPatch("/rest/api/v1/services/crossref", token, &body)

	docs.Then("Then the request is successful")
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, response.status)

	docs.Then("And the service has been correctly written, committed and pushed")
	filename := "owners/some-owner/services/crossref.yaml"
	expectedYaml := strings.ReplaceAll(tstServiceExpectedYaml("crossref"), "crossref/implementation", "not-crossref/implementation")
	require.Equal(t, expectedYaml, metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the service has been cached and can be read again, returning the correct information")
	readAgain, err := tstPerformGet("/rest/api/v1/services/crossref", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "service-patch-crossref.json")

	docs.Then("And a kafka message notifying other instances of the update has been sent")
	require.Equal(t, 4, len(kafkaImpl.Recording)) // the initial inserts also cause kafka messages
	actual, _ := json.Marshal(kafkaImpl.Recording[3])
	require.Equal(t, tstServiceExpectedKafka("crossref"), string(actual))
}

func TestPATCHService_ImplementationCrossref_ChangeOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.Given("Given existing repositories crossref.helm-deployment and not-crossref.implementation")
	deplRepoBody := tstRepository()
	deplRepoResponse, err := tstPerformPost("/rest/api/v1/repositories/crossref.helm-deployment", token, &deplRepoBody)
	require.Nil(t, err)
	require.Equal(t, http.StatusCreated, deplRepoResponse.status)

	implRepoBody := tstRepository()
	implRepoResponse, err := tstPerformPost("/rest/api/v1/repositories/not-crossref.implementation", token, &implRepoBody)
	require.Nil(t, err)
	require.Equal(t, http.StatusCreated, implRepoResponse.status)

	docs.Given("Given an existing service that references both these repositories, one of which is a cross reference")
	createBody := tstService("crossref")
	createBody.Repositories = []string{
		"crossref.helm-deployment",
		"not-crossref.implementation",
	}
	createResponse, err := tstPerformPost("/rest/api/v1/services/crossref", token, &createBody)
	require.Nil(t, err)
	require.Equal(t, http.StatusCreated, createResponse.status)
	// need commit hash from response, so we can perform valid update
	var tmp openapi.ServiceDto
	err = json.Unmarshal([]byte(createResponse.body), &tmp)
	require.Nil(t, err)

	docs.When("When they perform a valid patch of the service that moves it to a new owner")
	body := tstServicePatch()
	body.Owner = ptr("deleteme")
	body.CommitHash = tmp.CommitHash
	response, err := tstPerformPatch("/rest/api/v1/services/crossref", token, &body)

	docs.Then("Then the request is successful")
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, response.status)

	docs.Then("And the service with its referenced repositories has been correctly moved, committed and pushed, including the cross-referenced repository")
	filename1old := "owners/some-owner/services/crossref.yaml"
	filename1 := "owners/deleteme/services/crossref.yaml"
	filename2old := "owners/some-owner/repositories/crossref.helm-deployment.yaml"
	filename2 := "owners/deleteme/repositories/crossref.helm-deployment.yaml"
	filename3old := "owners/some-owner/repositories/not-crossref.implementation.yaml"
	filename3 := "owners/deleteme/repositories/not-crossref.implementation.yaml"
	expectedYaml := strings.ReplaceAll(tstServiceExpectedYaml("crossref"), "crossref/implementation", "not-crossref/implementation")
	require.Equal(t, expectedYaml, metadataImpl.ReadContents(filename1))
	require.True(t, metadataImpl.FilesCommitted[filename1])
	require.True(t, metadataImpl.FilesCommitted[filename1old])
	require.True(t, metadataImpl.FilesCommitted[filename2])
	require.True(t, metadataImpl.FilesCommitted[filename2old])
	require.True(t, metadataImpl.FilesCommitted[filename3])
	require.True(t, metadataImpl.FilesCommitted[filename3old])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the service has been cached and can be read again, returning the correct owner")
	readAgain, err := tstPerformGet("/rest/api/v1/services/crossref", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "service-patch-crossref-ownerchange.json")

	docs.Then("And a kafka message notifying other instances of the update has been sent")
	require.Equal(t, 4, len(kafkaImpl.Recording)) // the initial inserts also cause kafka messages
	actual, _ := json.Marshal(kafkaImpl.Recording[3])
	expectedMsg := strings.ReplaceAll(tstServiceMovedExpectedKafka("crossref"), "crossref.implementation", "not-crossref.implementation")
	require.Equal(t, expectedMsg, string(actual))
}

// delete service

func TestDELETEService_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they delete an existing service")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssertNoBody(t, response, err, http.StatusNoContent)

	docs.Then("And the service has been correctly deleted, committed and pushed")
	filename := "owners/some-owner/services/some-service-backend.yaml"
	require.Equal(t, "<notfound>", metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And its referenced repositories are left untouched (because otherwise they'd fall out of bit-brother control)")
	filename2 := "owners/some-owner/repositories/some-service-backend.helm-deployment.yaml"
	filename3 := "owners/some-owner/repositories/some-service-backend.helm-deployment.yaml"
	require.False(t, metadataImpl.FilesCommitted[filename2])
	require.False(t, metadataImpl.FilesCommitted[filename3])

	docs.Then("And the service has been removed from the cache")
	readAgain, err := tstPerformGet("/rest/api/v1/services/some-service-backend", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusNotFound, "service-notfound-someservicebackend.json")

	docs.Then("And a kafka message notifying other instances of the deletion has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstServiceExpectedKafka("some-service-backend"), string(actual))

	docs.Then("And a notification has been sent to all matching owners")
	hasSentNotification(t, "receivesDelete", "some-service-backend", types.DeletedEvent, types.ServicePayload, nil)
	hasSentNotification(t, "receivesService", "some-service-backend", types.DeletedEvent, types.ServicePayload, nil)
}

func TestDELETEService_DoesNotExist(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to delete a service that does not exist")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/services/does-not-exist", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusNotFound, "service-notfound-doesnotexist.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestDELETEService_MissingBody(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to delete a service but fail to provide a jira issue in the body")
	response, err := tstPerformDeleteNoBody("/rest/api/v1/services/some-service-backend", token)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-delete-invalid-values.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETEService_Unauthenticated(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user (no token)")
	token := tstUnauthenticated()

	docs.When("When they attempt to delete a service")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETEService_ExpiredToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with an expired token")
	token := tstExpiredAdminToken()

	docs.When("When they request deletion of a service")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETEService_NonAdminToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with a valid token without the admin role")
	token := tstValidUserToken()

	docs.When("When they attempt to delete a service")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusForbidden, "forbidden.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETEService_GitServerDown(t *testing.T) {
	tstReset()

	docs.Given("Given the git server is down")
	metadataImpl.SimulateRemoteFailure = true

	docs.Given("And given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request to delete a service")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadGateway, "bad-gateway.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETEService_GitHookDeclined(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request to delete a service, but supply an invalid issue")
	body := tstDelete()
	body.JiraIssue = "INVALID-12345"
	response, err := tstPerformDelete("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "receive-hook-declined.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

// get service promoters

func TestGETServicePromoters_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request the promoters for an existing service")
	response, err := tstPerformGet("/rest/api/v1/services/some-service-backend/promoters", token)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "service-promoters.json")
}

func TestGETServicePromoters_NotFound(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request the promoters for a service that does not exist")
	response, err := tstPerformGet("/rest/api/v1/services/unicorn/promoters", token)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusNotFound, "service-notfound.json")
}
