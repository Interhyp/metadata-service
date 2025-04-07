package acceptance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Interhyp/go-backend-service-common/docs"
	"github.com/go-http-utils/headers"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestPOSTWebhookGitHub_CheckSuite_Success(t *testing.T) {
	tstReset()

	docs.When("When GitHub sends a webhook with valid payload")
	body := createGithubCheckSuitePayload("a800c51995d3f3ee0ca110fa5fd93a772eaff381")

	bodyBytes, err := json.Marshal(&body)
	require.Nil(t, err)
	request, err := http.NewRequest(http.MethodPost, ts.URL+"/webhooks/vcs/github", bytes.NewReader(bodyBytes))
	require.Nil(t, err)
	request.Header.Set("X-GitHub-Event", string(github.CheckSuiteEvent))
	request.Header.Set(headers.ContentType, "application/json")
	rawResponse, err := http.DefaultClient.Do(request)
	require.Nil(t, err)
	response, err := tstWebResponseFromResponse(rawResponse)
	require.Nil(t, err)

	docs.Then("Then the request is successful")
	tstAssertNoBody(t, response, err, http.StatusNoContent)
}

func TestPOSTWebhookGitHub_InvalidCheckSuitePayload(t *testing.T) {
	tstReset()

	docs.When("When they send a webhook with invalid payload")
	request, err := http.NewRequest(http.MethodPost, ts.URL+"/webhooks/vcs/github", bytes.NewReader([]byte("")))
	require.Nil(t, err)
	request.Header.Set("X-GitHub-Event", string(github.CheckSuiteEvent))
	request.Header.Set(headers.ContentType, "application/json")
	rawResponse, err := http.DefaultClient.Do(request)
	require.Nil(t, err)
	response, err := tstWebResponseFromResponse(rawResponse)
	require.Nil(t, err)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "webhook-invalid.json")
}

func createGithubCheckSuitePayload(sha string) github.CheckSuitePayload {
	s := fmt.Sprintf(`{"action": "requested", "check_suite": {"head_sha": "%s", "app": {"created_at": "2025-04-04T00:00:00Z","updated_at": "2025-04-04T00:00:00Z"}}, "repository": {"name": "some-repo", "ssh_url": "ssh://git@github.com:Someorg/some-service-deployment.git", "owner": {"login": "some-org"}}}`, sha)
	data := github.CheckSuitePayload{}
	_ = json.Unmarshal([]byte(s), &data)
	return data
}
