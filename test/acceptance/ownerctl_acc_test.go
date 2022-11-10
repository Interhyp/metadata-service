package acceptance

import (
	"encoding/json"
	"github.com/Interhyp/metadata-service/docs"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

// get owners

func TestGETOwners_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request the list of owners")
	response, err := tstPerformGet("/rest/api/v1/owners", token)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "owners.json")
}

// get owner

func TestGETOwner_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request a single existing owner")
	response, err := tstPerformGet("/rest/api/v1/owners/some-owner", token)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "owner.json")
}

func TestGETOwner_NotFound(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they request a single owner that does not exist")
	response, err := tstPerformGet("/rest/api/v1/owners/migration-excellence", token)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusNotFound, "owner-notfound.json")
}

// create owner

func TestPOSTOwner_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a valid owner that does not exist")
	body := tstOwner()
	response, err := tstPerformPost("/rest/api/v1/owners/post-owner-success", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusCreated, "owner-create.json")

	docs.Then("And the owner has been correctly written, committed and pushed")
	filename := "owners/post-owner-success/owner.info.yaml"
	require.Equal(t, tstOwnerExpectedYaml(), metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the owner has been cached and can be read again")
	readAgain, err := tstPerformGet("/rest/api/v1/owners/post-owner-success", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "owner-create.json")

	docs.Then("And a kafka message notifying other instances of the creation has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstOwnerExpectedKafka("post-owner-success"), string(actual))
}

func TestPOSTOwner_InvalidAlias(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of an owner with an invalid alias")

	body := tstOwner()
	response, err := tstPerformPost("/rest/api/v1/owners/CapitalsAreForbidden", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "owner-invalid-alias.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTOwner_InvalidBody(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of an owner with an invalid body")
	body := []byte(`{"invalid json syntax":{,{`)
	response, err := tstPerformRawWithBody(http.MethodPost, "/rest/api/v1/owners/post-owner-invalid-syntax", token, body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "owner-invalid-syntax.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTOwner_InvalidValues(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of an owner with invalid values in the body")
	body := tstOwner()
	body.Contact = "" // not allowed
	response, err := tstPerformPost("/rest/api/v1/owners/post-owner-invalid-values", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "owner-create-invalid-values.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTOwner_Unauthenticated(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user (no token)")
	token := tstUnauthenticated()

	docs.When("When they request the creation of a valid owner that does not exist")
	body := tstOwner()
	response, err := tstPerformPost("/rest/api/v1/owners/post-owner-no-token", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTOwner_InvalidSignatureToken(t *testing.T) {
	tstReset()

	docs.Given("Given an attacker who tries to forge a token")
	token := tstInvalidSignatureToken()

	docs.When("When they request the creation of a valid owner that does not exist")
	body := tstOwner()
	response, err := tstPerformPost("/rest/api/v1/owners/post-owner-invalid-signature-token", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTOwner_ExpiredToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with an expired token")
	token := tstExpiredAdminToken()

	docs.When("When they request the creation of a valid owner that does not exist")
	body := tstOwner()
	response, err := tstPerformPost("/rest/api/v1/owners/post-owner-expired-token", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTOwner_NonAdminToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with a valid token without the admin role")
	token := tstValidUserToken()

	docs.When("When they request the creation of a valid owner that does not exist")
	body := tstOwner()
	response, err := tstPerformPost("/rest/api/v1/owners/post-owner-user-token", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusForbidden, "forbidden.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTOwner_Duplicate(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a valid owner that already exists")
	body := tstOwner()
	response, err := tstPerformPost("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusConflict, "owner-create-duplicate.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPOSTOwner_GitServerDown(t *testing.T) {
	tstReset()

	docs.Given("Given the git server is down")
	metadataImpl.SimulateRemoteFailure = true

	docs.Given("And given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request the creation of a valid owner")
	body := tstOwner()
	response, err := tstPerformPost("/rest/api/v1/owners/post-owner-git-down", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadGateway, "bad-gateway.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

// update full owner

func TestPUTOwner_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid update of an existing owner")
	body := tstOwner()
	response, err := tstPerformPut("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "owner-update.json")

	docs.Then("And the owner has been correctly written, committed and pushed")
	filename := "owners/some-owner/owner.info.yaml"
	require.Equal(t, tstOwnerExpectedYaml(), metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the owner has been cached and can be read again")
	readAgain, err := tstPerformGet("/rest/api/v1/owners/some-owner", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "owner-update.json")

	docs.Then("And a kafka message notifying other instances of the update has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstOwnerExpectedKafka("some-owner"), string(actual))
}

func TestPUTOwner_NoChangeSuccess(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid update of an existing owner that does not make actual changes")
	metadataImpl.SimulateUnchangedFailure = true
	body := tstOwnerUnchanged()
	response, err := tstPerformPut("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "owner-unchanged.json")

	docs.Then("And no commit has been made and pushed (because there were no changes)")
	filename := "owners/some-owner/owner.info.yaml"
	require.Equal(t, tstOwnerUnchangedExpectedYaml(), metadataImpl.ReadContents(filename))
	require.False(t, metadataImpl.FilesCommitted[filename])
	require.False(t, metadataImpl.Pushed)

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestPUTOwner_DoesNotExist(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt a valid update of an owner that does not exist")
	body := tstOwner()
	response, err := tstPerformPut("/rest/api/v1/owners/does-not-exist", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusNotFound, "owner-notfound.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestPUTOwner_InvalidBody(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of an owner with a syntactically invalid body")
	body := []byte(`{"invalid json syntax":{,{`)
	response, err := tstPerformRawWithBody(http.MethodPut, "/rest/api/v1/owners/some-owner", token, body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "owner-invalid-syntax.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTOwner_InvalidValues(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of an owner with invalid values in the body")
	body := tstOwner()
	body.Contact = ""
	body.CommitHash = ""
	body.TimeStamp = ""
	response, err := tstPerformPut("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "owner-update-invalid-values.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTOwner_Unauthenticated(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user (no token)")
	token := tstUnauthenticated()

	docs.When("When they request an update of an owner")
	body := tstOwner()
	response, err := tstPerformPut("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTOwner_InvalidAlgorithmToken(t *testing.T) {
	tstReset()

	docs.Given("Given an attacker who tries a token algorithm attack")
	token := tstInvalidAlgorithmToken()

	docs.When("When they request an update of an owner")
	body := tstOwner()
	response, err := tstPerformPut("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTOwner_ExpiredToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with an expired token")
	token := tstExpiredAdminToken()

	docs.When("When they request an update of an owner")
	body := tstOwner()
	response, err := tstPerformPut("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTOwner_NonAdminToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with a valid token without the admin role")
	token := tstValidUserToken()

	docs.When("When they request an update of an owner")
	body := tstOwner()
	response, err := tstPerformPut("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusForbidden, "forbidden.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTOwner_Conflict(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of an owner based on an outdated commit hash")
	body := tstOwner()
	body.CommitHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	body.TimeStamp = "2019-01-01T00:00:00Z"
	response, err := tstPerformPut("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusConflict, "owner-update-conflict.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPUTOwner_GitServerDown(t *testing.T) {
	tstReset()

	docs.Given("Given the git server is down")
	metadataImpl.SimulateRemoteFailure = true

	docs.Given("And given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request an update of an owner")
	body := tstOwner()
	response, err := tstPerformPut("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadGateway, "bad-gateway.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

// patch owner

func TestPATCHOwner_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid patch of an existing owner")
	body := tstOwnerPatch()
	response, err := tstPerformPatch("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "owner-patch.json")

	docs.Then("And the owner has been correctly written, committed and pushed")
	filename := "owners/some-owner/owner.info.yaml"
	require.Equal(t, tstOwnerPatchExpectedYaml(), metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the owner has been cached and can be read again")
	readAgain, err := tstPerformGet("/rest/api/v1/owners/some-owner", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusOK, "owner-patch.json")

	docs.Then("And a kafka message notifying other instances of the update has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstOwnerExpectedKafka("some-owner"), string(actual))
}

func TestPATCHOwner_NoChangeSuccess(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they perform a valid patch of an existing owner that does not make actual changes")
	metadataImpl.SimulateUnchangedFailure = true
	body := tstOwnerUnchangedPatch()
	response, err := tstPerformPatch("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssert(t, response, err, http.StatusOK, "owner-unchanged.json")

	docs.Then("And no commit has been made and pushed (because there were no changes)")
	filename := "owners/some-owner/owner.info.yaml"
	require.Equal(t, tstOwnerUnchangedExpectedYaml(), metadataImpl.ReadContents(filename))
	require.False(t, metadataImpl.FilesCommitted[filename])
	require.False(t, metadataImpl.Pushed)

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestPATCHOwner_DoesNotExist(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to patch an owner that does not exist")
	body := tstOwnerPatch()
	response, err := tstPerformPatch("/rest/api/v1/owners/does-not-exist", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusNotFound, "owner-notfound.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestPATCHOwner_InvalidBody(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request a patch of an owner with a syntactically invalid body")
	body := []byte(`},},},},},},},}`)
	response, err := tstPerformRawWithBody(http.MethodPatch, "/rest/api/v1/owners/some-owner", token, body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "owner-invalid-syntax.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHOwner_InvalidValues(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request a patch of an owner with invalid values in the body")
	body := tstOwnerPatch()
	body.Contact = p("") // invalid
	body.CommitHash = ""
	body.TimeStamp = ""
	response, err := tstPerformPatch("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "owner-patch-invalid-values.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHOwner_Unauthenticated(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user (no token)")
	token := tstUnauthenticated()

	docs.When("When they request a patch of an owner")
	body := tstOwnerPatch()
	response, err := tstPerformPatch("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHOwner_ExpiredToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with an expired token")
	token := tstExpiredAdminToken()

	docs.When("When they request a patch of an owner")
	body := tstOwnerPatch()
	response, err := tstPerformPatch("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHOwner_NonAdminToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with a valid token without the admin role")
	token := tstValidUserToken()

	docs.When("When they request a patch of an owner")
	body := tstOwnerPatch()
	response, err := tstPerformPatch("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusForbidden, "forbidden.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHOwner_Conflict(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request a patch of an owner based on an outdated commit hash")
	body := tstOwnerPatch()
	body.CommitHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	body.TimeStamp = "2019-01-01T00:00:00Z"
	response, err := tstPerformPatch("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusConflict, "owner-patch-conflict.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestPATCHOwner_GitServerDown(t *testing.T) {
	tstReset()

	docs.Given("Given the git server is down")
	metadataImpl.SimulateRemoteFailure = true

	docs.Given("And given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request a patch of an owner")
	body := tstOwnerPatch()
	response, err := tstPerformPatch("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadGateway, "bad-gateway.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

// delete owner

func TestDELETEOwner_Success(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they delete an existing owner that has no services and repositories")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/owners/deleteme", token, &body)

	docs.Then("Then the request is successful and the response is as expected")
	tstAssertNoBody(t, response, err, http.StatusNoContent)

	docs.Then("And the owner has been correctly deleted, committed and pushed")
	filename := "owners/deleteme/owner.info.yaml"
	require.Equal(t, "<notfound>", metadataImpl.ReadContents(filename))
	require.True(t, metadataImpl.FilesCommitted[filename])
	require.True(t, metadataImpl.Pushed)

	docs.Then("And the owner has been removed from the cache")
	readAgain, err := tstPerformGet("/rest/api/v1/owners/deleteme", tstUnauthenticated())
	tstAssert(t, readAgain, err, http.StatusNotFound, "owner-notfound.json")

	docs.Then("And a kafka message notifying other instances of the deletion has been sent")
	require.Equal(t, 1, len(kafkaImpl.Recording))
	actual, _ := json.Marshal(kafkaImpl.Recording[0])
	require.Equal(t, tstOwnerExpectedKafka("deleteme"), string(actual))
}

func TestDELETEOwner_DoesNotExist(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to delete an owner that does not exist")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/owners/does-not-exist", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusNotFound, "owner-notfound.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))

	docs.Then("And no kafka messages have been sent")
	require.Equal(t, 0, len(kafkaImpl.Recording))
}

func TestDELETEOwner_MissingBody(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they attempt to delete an owner but fail to provide a jira issue in the body")
	response, err := tstPerformDeleteNoBody("/rest/api/v1/owners/deleteme", token)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "owner-delete-invalid-values.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETEOwner_Unauthenticated(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user (no token)")
	token := tstUnauthenticated()

	docs.When("When they attempt to delete an owner")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETEOwner_ExpiredToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with an expired token")
	token := tstExpiredAdminToken()

	docs.When("When they request deletion of an owner")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusUnauthorized, "unauthorized.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETEOwner_NonAdminToken(t *testing.T) {
	tstReset()

	docs.Given("Given a user with a valid token without the admin role")
	token := tstValidUserToken()

	docs.When("When they attempt to delete an owner")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/owners/deleteme", token, &body)

	docs.Then("Then the request is denied")
	tstAssert(t, response, err, http.StatusForbidden, "forbidden.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETEOwner_StillHasStuffConflict(t *testing.T) {
	tstReset()

	docs.Given("Given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request to delete an owner that still owns a service")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/owners/some-owner", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusConflict, "owner-delete-conflict.json")

	docs.Then("And no changes have been made in the metadata repository")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}

func TestDELETEOwner_GitServerDown(t *testing.T) {
	tstReset()

	docs.Given("Given the git server is down")
	metadataImpl.SimulateRemoteFailure = true

	docs.Given("And given an authenticated admin user")
	token := tstValidAdminToken()

	docs.When("When they request to delete an owner")
	body := tstDelete()
	response, err := tstPerformDelete("/rest/api/v1/owners/deleteme", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadGateway, "bad-gateway.json")

	docs.Then("And the local metadata repository clone has been reset to its original state")
	require.Equal(t, 0, len(metadataImpl.FilesWritten))
	require.Equal(t, 0, len(metadataImpl.FilesCommitted))
}
