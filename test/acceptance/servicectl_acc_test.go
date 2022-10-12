package acceptance

import (
	"encoding/json"
	"github.com/Interhyp/metadata-service/docs"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
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

func TestGETService_InvalidAlias(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request a single service with an invalid name")
	response, err := tstPerformGet("/rest/api/v1/services/ääääää", token)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-invalid.json")
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
	body := tstService("new-service")
	response, err := tstPerformPost("/rest/api/v1/services/new-service", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusCreated, "service-create.json")

	docs.Then("And the service has been correctly written, committed and pushed")
	filename := "owners/some-owner/services/new-service.yaml"
	require.Equal(t, tstServiceExpectedYaml("new-service"), metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the service has been cached and can be read again")
	readAgain, err := tstPerformGet("/rest/api/v1/services/new-service", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "service-create.json")

	docs.Then("And a kafka message notifying other instances of the creation has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstServiceExpectedKafka("new-service"), string(actual))
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
		"new-service.helm-deployment", // cross ref to other service not allowed
	}
	response, err := tstPerformPost("/rest/api/v1/services/post-service-invalid-values", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-create-invalid-values.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTService_NonexistantRepository(t *testing.T) {
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

func TestPOSTService_NonexistantOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a service referring to an owner that does not exist")
	body := tstService("new-service")
	body.Owner = "not-there"
	response, err := tstPerformPost("/rest/api/v1/services/new-service", token, &body)

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
	body := tstService("new-service")
	response, err := tstPerformPost("/rest/api/v1/services/new-service", token, &body)

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
	body := tstService("new-service")
	response, err := tstPerformPost("/rest/api/v1/services/new-service", token, &body)

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
	body := tstService("new-service")
	response, err := tstPerformPost("/rest/api/v1/services/new-service", token, &body)

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
	body := tstService("new-service")
	response, err := tstPerformPost("/rest/api/v1/services/new-service", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadGateway, "bad-gateway.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
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
	tstAssert(t, response, err, http.StatusNotFound, "service-notfound.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestPUTService_NonexistantRepository(t *testing.T) {
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

func TestPUTService_NonexistantOwner(t *testing.T) {
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

func TestPUTService_InvalidName(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a service with an invalid name")

	body := tstService("abc")
	response, err := tstPerformPut("/rest/api/v1/services/ABC'", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-invalid-name.json")

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
	tstAssert(t, response, err, http.StatusNotFound, "service-notfound.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestPATCHService_NonexistantRepository(t *testing.T) {
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

func TestPATCHService_NonexistantOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to patch a service referring to an owner that does not exist")
	body := tstServicePatch()
	body.Owner = p("not-there")
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-update-owner-missing.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHService_InvalidName(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to patch a service with an invalid name")

	body := tstServicePatch()
	response, err := tstPerformPatch("/rest/api/v1/services/INVALID'", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-invalid-name.json")

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
	body.Owner = p("")       // invalid
	body.AlertTarget = p("") // invalid (nil would be valid)
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

func TestPATCHService_ChangeOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid patch of an existing service that changes its owner")
	body := tstServicePatch()
	body.Owner = p("deleteme")
	response, err := tstPerformPatch("/rest/api/v1/services/some-service-backend", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "service-patch-newowner.json")

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
	tstAssert(t, readAgain, err, http.StatusOK, "service-patch-newowner.json")

	docs.Then("And a kafka message notifying other instances of the update has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstServiceMovedExpectedKafka("some-service-backend"), string(actual))
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
	tstAssert(t, readAgain, err, http.StatusNotFound, "service-notfound.json")

	docs.Then("And a kafka message notifying other instances of the deletion has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstServiceExpectedKafka("some-service-backend"), string(actual))
}

func TestDELETEService_DoesNotExist(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to delete a service that does not exist")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/services/does-not-exist", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusNotFound, "service-notfound.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestDELETEService_InvalidName(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to delete a service with an invalid name")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/services/___'", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-invalid-name.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
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

func TestGETServicePromoters_InvalidName(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request the promoters for a service with an invalid name")
	response, err := tstPerformGet("/rest/api/v1/services/äbü/promoters", token)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "service-invalid.json")
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

func TestGETServicePromoters_Success_WithAdditionals(t *testing.T) {
	tstReset()

	docs.Given("Given extra promoters have been added to the owner configured as additional promoters source")
	body := tstOwnerPatch()
	ownerPatchResponse, err := tstPerformPatch("/rest/api/v1/owners/deleteme", tstValidAdminToken(), &body)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, ownerPatchResponse.status)

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request the promoters for an existing service")
	response, err := tstPerformGet("/rest/api/v1/services/some-service-backend/promoters", token)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "service-promoters-with-additionals.json")
}
