package acceptance

import (
	"encoding/json"
	"github.com/Interhyp/metadata-service/internal/types"
	"github.com/StephanHCB/go-backend-service-common/docs"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

// get repositories

func TestGETRepositories_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request the list of repositories")
	response, err := tstPerformGet("/rest/api/v1/repositories", token)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "repositories.json")
}

func TestGETRepositories_Filtered_NonexistingService(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request the list of repositories filtered by a service that does not exist")
	response, err := tstPerformGet("/rest/api/v1/repositories?service=does-not-exist", token)

	docs.Then("Then the request is successful and the response contains an empty result")
	tstAssert(t, response, err, http.StatusOK, "repositories-empty.json")
}

func TestGETRepositories_Filtered_NonexistingOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request the list of repositories filtered by an owner that does not exist")
	response, err := tstPerformGet("/rest/api/v1/repositories?owner=does-not-exist", token)

	docs.Then("Then the request is successful and the response contains an empty result")
	tstAssert(t, response, err, http.StatusOK, "repositories-empty.json")
}

func TestGETRepositories_Filtered_Service(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request the list of repositories filtered by a single service")
	response, err := tstPerformGet("/rest/api/v1/repositories?service=some-service-backend", token)

	docs.Then("Then the request is successful and the response contains an empty result")
	tstAssert(t, response, err, http.StatusOK, "repositories-filtered-service.json")
}

func TestGETRepositories_Filtered_Type(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request the list of repositories filtered by a single service")
	response, err := tstPerformGet("/rest/api/v1/repositories?type=implementation", token)

	docs.Then("Then the request is successful and the response contains an empty result")
	tstAssert(t, response, err, http.StatusOK, "repositories-filtered-type.json")
}

// get repository

func TestGETRepository_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request a single existing repository")
	response, err := tstPerformGet("/rest/api/v1/repositories/some-service-backend.helm-deployment", token)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "repository.json")
}

func TestGETRepository_NotFound(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request a single repository that does not exist")
	response, err := tstPerformGet("/rest/api/v1/repositories/unicorn.helm-chart", token)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusNotFound, "repository-notfound.json")
}

// create repository

func TestPOSTRepository_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a valid repository that does not exist")
	body := tstRepository()
	response, err := tstPerformPost("/rest/api/v1/repositories/new-repository.api", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusCreated, "repository-create.json")

	docs.Then("And the repository has been correctly written, committed and pushed")
	filename := "owners/some-owner/repositories/new-repository.api.yaml"
	require.Equal(t, tstRepositoryExpectedYaml(), metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the repository has been cached and can be read again")
	readAgain, err := tstPerformGet("/rest/api/v1/repositories/new-repository.api", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "repository-create.json")

	docs.Then("And a kafka message notifying other instances of the creation has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstRepositoryExpectedKafka("new-repository.api"), string(actual))

	docs.Then("And a notification has been sent to all matching owners")
	payload := tstNewRepositoryPayload()
	hasSentNotification(t, "receivesCreate", "new-repository.api", types.CreatedEvent, types.RepositoryPayload, &payload)
	hasSentNotification(t, "receivesRepository", "new-repository.api", types.CreatedEvent, types.RepositoryPayload, &payload)
}

func TestPOSTRepository_InvalidKey(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a repository with an invalid key")

	body := tstRepository()
	response, err := tstPerformPost("/rest/api/v1/repositories/-ab.wrong", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "repository-invalid-key.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTRepository_InvalidBody(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a repository with an invalid body")
	body := []byte(`{"invalid json syntax":{,{`)
	response, err := tstPerformRawWithBody(http.MethodPost, "/rest/api/v1/repositories/post-repository-invalid-syntax.api", token, body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "repository-invalid-syntax.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTRepository_InvalidValues(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a repository with invalid values in the body")
	body := tstRepository()
	body.Owner = ""
	response, err := tstPerformPost("/rest/api/v1/repositories/post-repository-invalid-values.api", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "repository-create-invalid-values.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTRepository_NonexistentOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a repository referring to an owner that does not exist")
	body := tstRepository()
	body.Owner = "not-there"
	response, err := tstPerformPost("/rest/api/v1/repositories/new-repository.helm-chart", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "repository-create-owner-missing.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTRepository_Unauthenticated(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user (no token)")
	token := tstUnauthenticated()

	docs.When("When they request the creation of a valid repository that does not exist")
	body := tstRepository()
	response, err := tstPerformPost("/rest/api/v1/repositories/new-repository.api", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTRepository_ExpiredToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with an expired token")
	token := tstExpiredAdminToken()

	docs.When("When they request the creation of a valid repository that does not exist")
	body := tstRepository()
	response, err := tstPerformPost("/rest/api/v1/repositories/new-repository.api", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTRepository_NonAdminToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with a valid token without the admin role")
	token := tstValidUserToken()

	docs.When("When they request the creation of a valid repository that does not exist")
	body := tstRepository()
	response, err := tstPerformPost("/rest/api/v1/repositories/new-repository.api", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusForbidden, "forbidden.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTRepository_Duplicate(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a valid repository that already exists")
	body := tstRepository()
	response, err := tstPerformPost("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusConflict, "repository-create-duplicate.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTRepository_GitServerDown(t *testing.T) {
	tstReset()

	docs.Given("Given the git server is down")
	metadataImpl.SimulateRemoteFailure = true

	docs.Given("And given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a valid repository")
	body := tstRepository()
	response, err := tstPerformPost("/rest/api/v1/repositories/new-repository.api", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadGateway, "bad-gateway.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTRepository_GitHookDeclined(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a valid repository, but supply an invalid issue")
	body := tstRepository()
	body.JiraIssue = "INVALID-12345"
	response, err := tstPerformPost("/rest/api/v1/repositories/new-repository.api", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "receive-hook-declined.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

// update full repository

func TestPUTRepository_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid update of an existing repository")
	body := tstRepository()
	response, err := tstPerformPut("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "repository-update.json")

	docs.Then("And the repository has been correctly written, committed and pushed")
	filename := "owners/some-owner/repositories/karma-wrapper.helm-chart.yaml"
	require.Equal(t, tstRepositoryExpectedYaml(), metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the repository has been cached and can be read again")
	readAgain, err := tstPerformGet("/rest/api/v1/repositories/karma-wrapper.helm-chart", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "repository-update.json")

	docs.Then("And a kafka message notifying other instances of the update has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstRepositoryExpectedKafka("karma-wrapper.helm-chart"), string(actual))

	docs.Then("And a notification has been sent to all matching owners")
	payload := tstNewRepositoryPayload()
	hasSentNotification(t, "receivesModified", "karma-wrapper.helm-chart", types.ModifiedEvent, types.RepositoryPayload, &payload)
	hasSentNotification(t, "receivesRepository", "karma-wrapper.helm-chart", types.ModifiedEvent, types.RepositoryPayload, &payload)
}

func TestPUTRepository_NoChangeSuccess(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid update of an existing repository that does not make actual changes")
	metadataImpl.SimulateUnchangedFailure = true
	body := tstRepositoryUnchanged()
	response, err := tstPerformPut("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "repository-unchanged.json")

	docs.Then("And no commit has been made and pushed (because there were no changes)")
	filename := "owners/some-owner/repositories/karma-wrapper.helm-chart.yaml"
	require.Equal(t, tstRepositoryUnchangedExpectedYaml(), metadataImpl.ReadContents(filename))
	require.False(t, metadataImpl.FilesCommitted[filename])
	require.False(t, metadataImpl.Pushed)

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestPUTRepository_DoesNotExist(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt a valid update of a repository that does not exist")
	body := tstRepository()
	response, err := tstPerformPut("/rest/api/v1/repositories/does-not-exist.api", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusNotFound, "repository-notfound-doesnotexist.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestPUTRepository_NonexistentOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a repository referring to an owner that does not exist")
	body := tstRepository()
	body.Owner = "not-there"
	response, err := tstPerformPut("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "repository-update-owner-missing.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTRepository_InvalidBody(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a repository with a syntactically invalid body")
	body := []byte(`{"invalid json syntax":{,{`)
	response, err := tstPerformRawWithBody(http.MethodPut, "/rest/api/v1/repositories/karma-wrapper.helm-chart", token, body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "repository-invalid-syntax.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTRepository_InvalidValues(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a repository with invalid values in the body")
	body := tstRepository()
	body.Url = ""
	body.CommitHash = ""
	body.TimeStamp = ""
	response, err := tstPerformPut("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "repository-update-invalid-values.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTRepository_Unauthenticated(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user (no token)")
	token := tstUnauthenticated()

	docs.When("When they request an update of a repository")
	body := tstRepository()
	response, err := tstPerformPut("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTRepository_ExpiredToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with an expired token")
	token := tstExpiredAdminToken()

	docs.When("When they request an update of a repository")
	body := tstRepository()
	response, err := tstPerformPut("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTRepository_NonAdminToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with a valid token without the admin role")
	token := tstValidUserToken()

	docs.When("When they request an update of a repository")
	body := tstRepository()
	response, err := tstPerformPut("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusForbidden, "forbidden.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTRepository_Conflict(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a repository based on an outdated commit hash")
	body := tstRepository()
	body.CommitHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	body.TimeStamp = "2019-01-01T00:00:00Z"
	response, err := tstPerformPut("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusConflict, "repository-update-conflict.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTRepository_GitServerDown(t *testing.T) {
	tstReset()

	docs.Given("Given the git server is down")
	metadataImpl.SimulateRemoteFailure = true

	docs.Given("And given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a repository")
	body := tstRepository()
	response, err := tstPerformPut("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadGateway, "bad-gateway.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTRepository_GitHookDeclined(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of a repository, but supply an invalid issue")
	body := tstRepository()
	body.JiraIssue = "INVALID-12345"
	response, err := tstPerformPut("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "receive-hook-declined.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTRepository_ChangeOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid update of an existing repository that changes its owner")
	body := tstRepository()
	body.Owner = "deleteme"
	response, err := tstPerformPut("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "repository-update-newowner.json")

	docs.Then("And the repository with its repositories has been correctly moved, committed and pushed")
	filename1old := "owners/some-owner/repositories/karma-wrapper.helm-chart.yaml"
	filename1 := "owners/deleteme/repositories/karma-wrapper.helm-chart.yaml"
	require.Equal(t, tstRepositoryExpectedYaml(), metadataImpl.ReadContents(filename1))
	require.True(t, metadataImpl.FilesCommitted[filename1])
	require.True(t, metadataImpl.FilesCommitted[filename1old])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the repository has been cached and can be read again, returning the correct owner")
	readAgain, err := tstPerformGet("/rest/api/v1/repositories/karma-wrapper.helm-chart", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "repository-update-newowner.json")

	docs.Then("And a kafka message notifying other instances of the update has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstRepositoryExpectedKafka("karma-wrapper.helm-chart"), string(actual))
}

func TestPUTRepository_ChangeOwnerReferenced(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to change the owner of a repository that is referenced in its service")
	body := tstRepository()
	body.Owner = "deleteme"
	response, err := tstPerformPut("/rest/api/v1/repositories/some-service-backend.helm-deployment", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusConflict, "repository-update-referenced.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

// patch repository

func TestPATCHRepository_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid patch of an existing repository")
	body := tstRepositoryPatch()
	response, err := tstPerformPatch("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "repository-patch.json")

	docs.Then("And the repository has been correctly written, committed and pushed")
	filename := "owners/some-owner/repositories/karma-wrapper.helm-chart.yaml"
	require.Equal(t, tstRepositoryExpectedYamlKarmaWrapper(), metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the repository has been cached and can be read again")
	readAgain, err := tstPerformGet("/rest/api/v1/repositories/karma-wrapper.helm-chart", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "repository-patch.json")

	docs.Then("And a kafka message notifying other instances of the update has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstRepositoryExpectedKafka("karma-wrapper.helm-chart"), string(actual))

	docs.Then("And a notification has been sent to all matching owners")
	payload := tstUpdatedRepositoryPayload()
	hasSentNotification(t, "receivesModified", "karma-wrapper.helm-chart", types.ModifiedEvent, types.RepositoryPayload, &payload)
	hasSentNotification(t, "receivesRepository", "karma-wrapper.helm-chart", types.ModifiedEvent, types.RepositoryPayload, &payload)
}

func TestPATCHRepository_UnsupportedConfigurationFieldsAreIgnored(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a patch changing unsupported fields in the configuration of an existing repository")
	body := tstRepositoryPatchWithIgnoredConfigurationFields()
	response, err := tstPerformPatch("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "repository-patch.json")

	docs.Then("And the repository has been correctly written, committed and pushed")
	filename := "owners/some-owner/repositories/karma-wrapper.helm-chart.yaml"
	require.Equal(t, tstRepositoryExpectedYamlKarmaWrapper(), metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the repository has been cached and can be read again")
	readAgain, err := tstPerformGet("/rest/api/v1/repositories/karma-wrapper.helm-chart", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "repository-patch.json")

	docs.Then("And a kafka message notifying other instances of the update has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstRepositoryExpectedKafka("karma-wrapper.helm-chart"), string(actual))

	docs.Then("And a notification has been sent to all matching owners")
	payload := tstUpdatedRepositoryPayload()
	hasSentNotification(t, "receivesModified", "karma-wrapper.helm-chart", types.ModifiedEvent, types.RepositoryPayload, &payload)
	hasSentNotification(t, "receivesRepository", "karma-wrapper.helm-chart", types.ModifiedEvent, types.RepositoryPayload, &payload)
}

func TestPATCHRepository_NoChangeSuccess(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid patch of an existing repository that does not make actual changes")
	metadataImpl.SimulateUnchangedFailure = true
	body := tstRepositoryUnchangedPatch()
	response, err := tstPerformPatch("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "repository-unchanged.json")

	docs.Then("And no commit has been made and pushed (because there were no changes)")
	filename := "owners/some-owner/repositories/karma-wrapper.helm-chart.yaml"
	require.Equal(t, tstRepositoryUnchangedExpectedYaml(), metadataImpl.ReadContents(filename))
	require.False(t, metadataImpl.FilesCommitted[filename])
	require.False(t, metadataImpl.Pushed)

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestPATCHRepository_DoesNotExist(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to patch a repository that does not exist")
	body := tstRepositoryPatch()
	response, err := tstPerformPatch("/rest/api/v1/repositories/does-not-exist.api", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusNotFound, "repository-notfound-doesnotexist.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestPATCHRepository_NonexistentOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to patch a repository referring to an owner that does not exist")
	body := tstRepositoryPatch()
	body.Owner = ptr("not-there")
	response, err := tstPerformPatch("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "repository-update-owner-missing.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHRepository_InvalidBody(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request a patch of a repository with a syntactically invalid body")
	body := []byte(`},},},},},},},}`)
	response, err := tstPerformRawWithBody(http.MethodPatch, "/rest/api/v1/repositories/karma-wrapper.helm-chart", token, body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "repository-invalid-syntax.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHRepository_InvalidValues(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request a patch of a repository with invalid values in the body")
	body := tstRepositoryPatch()
	body.Owner = ptr("") // invalid
	body.CommitHash = ""
	body.TimeStamp = ""
	response, err := tstPerformPatch("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "repository-patch-invalid-values.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHRepository_Unauthenticated(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user (no token)")
	token := tstUnauthenticated()

	docs.When("When they attempt to patch a repository")
	body := tstRepositoryPatch()
	response, err := tstPerformPatch("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHRepository_ExpiredToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with an expired token")
	token := tstExpiredAdminToken()

	docs.When("When they attempt to patch a repository")
	body := tstRepositoryPatch()
	response, err := tstPerformPatch("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHRepository_NonAdminToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with a valid token without the admin role")
	token := tstValidUserToken()

	docs.When("When they attempt to patch a repository")
	body := tstRepositoryPatch()
	response, err := tstPerformPatch("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusForbidden, "forbidden.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHRepository_Conflict(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to patch a repository based on an outdated commit hash")
	body := tstRepositoryPatch()
	body.CommitHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	body.TimeStamp = "2019-01-01T00:00:00Z"
	response, err := tstPerformPatch("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusConflict, "repository-patch-conflict.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHRepository_GitServerDown(t *testing.T) {
	tstReset()

	docs.Given("Given the git server is down")
	metadataImpl.SimulateRemoteFailure = true

	docs.Given("And given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to patch a repository")
	body := tstRepositoryPatch()
	response, err := tstPerformPatch("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadGateway, "bad-gateway.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHRepository_GitHookDeclined(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to patch a repository, but supply an invalid issue")
	body := tstRepositoryPatch()
	body.JiraIssue = "INVALID-12345"
	response, err := tstPerformPatch("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "receive-hook-declined.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHRepository_ChangeOwner(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid patch of an existing repository that changes its owner")
	body := tstRepositoryUnchangedPatch()
	body.Owner = ptr("deleteme")
	response, err := tstPerformPatch("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "repository-patch-newowner.json")

	docs.Then("And the repository with its repositories has been correctly moved, committed and pushed")
	filename1old := "owners/some-owner/repositories/karma-wrapper.helm-chart.yaml"
	filename1 := "owners/deleteme/repositories/karma-wrapper.helm-chart.yaml"
	require.Equal(t, tstRepositoryUnchangedExpectedYaml(), metadataImpl.ReadContents(filename1))
	require.True(t, metadataImpl.FilesCommitted[filename1])
	require.True(t, metadataImpl.FilesCommitted[filename1old])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the repository has been cached and can be read again, returning the correct owner")
	readAgain, err := tstPerformGet("/rest/api/v1/repositories/karma-wrapper.helm-chart", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "repository-patch-newowner.json")

	docs.Then("And a kafka message notifying other instances of the update has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstRepositoryExpectedKafka("karma-wrapper.helm-chart"), string(actual))
}

func TestPATCHRepository_ChangeOwnerReferenced(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to change the owner of a repository that is referenced in its service")
	body := tstRepositoryPatch()
	body.Owner = ptr("deleteme")
	response, err := tstPerformPatch("/rest/api/v1/repositories/some-service-backend.helm-deployment", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusConflict, "repository-update-referenced.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

// delete repository

func TestDELETERepository_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they delete an existing repository")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssertNoBody(t, response, err, http.StatusNoContent)

	docs.Then("And the repository has been correctly deleted, committed and pushed")
	filename := "owners/some-owner/repositories/karma-wrapper.helm-chart.yaml"
	require.Equal(t, "<notfound>", metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the repository has been removed from the cache")
	readAgain, err := tstPerformGet("/rest/api/v1/repositories/karma-wrapper.helm-chart", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusNotFound, "repository-notfound-karmawrapper.json")

	docs.Then("And a kafka message notifying other instances of the deletion has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstRepositoryExpectedKafka("karma-wrapper.helm-chart"), string(actual))

	docs.Then("And a notification has been sent to all matching owners")
	hasSentNotification(t, "receivesDelete", "karma-wrapper.helm-chart", types.DeletedEvent, types.RepositoryPayload, nil)
	hasSentNotification(t, "receivesRepository", "karma-wrapper.helm-chart", types.DeletedEvent, types.RepositoryPayload, nil)
}

func TestDELETERepository_DoesNotExist(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to delete a repository that does not exist")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/repositories/does-not-exist.api", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusNotFound, "repository-notfound-doesnotexist.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestDELETERepository_MissingBody(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to delete a repository but fail to provide a jira issue in the body")
	response, err := tstPerformDeleteNoBody("/rest/api/v1/repositories/karma-wrapper.helm-chart", token)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "repository-delete-invalid-values.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETERepository_Unauthenticated(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user (no token)")
	token := tstUnauthenticated()

	docs.When("When they attempt to delete a repository")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETERepository_ExpiredToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with an expired token")
	token := tstExpiredAdminToken()

	docs.When("When they request deletion of a repository")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETERepository_NonAdminToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with a valid token without the admin role")
	token := tstValidUserToken()

	docs.When("When they attempt to delete a repository")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusForbidden, "forbidden.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETERepository_GitServerDown(t *testing.T) {
	tstReset()

	docs.Given("Given the git server is down")
	metadataImpl.SimulateRemoteFailure = true

	docs.Given("And given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request to delete a repository")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadGateway, "bad-gateway.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETERepository_GitHookDeclined(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request to delete a repository, but supply an invalid issue")
	body := tstDelete()
	body.JiraIssue = "INVALID-12345"
	response, err := tstPerformDelete("/rest/api/v1/repositories/karma-wrapper.helm-chart", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "receive-hook-declined.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETERepository_Referenced(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to delete a repository that is referenced in its service")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/repositories/some-service-backend.helm-deployment", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusConflict, "repository-delete-referenced.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}
