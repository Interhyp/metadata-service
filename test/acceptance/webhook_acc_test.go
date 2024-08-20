package acceptance

import (
	"bytes"
	"encoding/json"
	"github.com/Interhyp/go-backend-service-common/docs"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	bitbucketserver "github.com/go-playground/webhooks/v6/bitbucket-server"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestPOSTWebhookBitbucket_Success(t *testing.T) {
	tstReset()

	docs.Given("Given a pull request in BitBucket with valid file edits")
	bbImpl.PRHead = "e2d2000000000000000000000000000000000000"
	bbImpl.ChangedFilesResponse = []repository.File{
		{
			Path: "owners/fun-owner/owner.info.yaml",
			Contents: `contact: someone@example.com
productOwner: someone
displayName: Test Owner`,
		},
		{
			Path: "owners/fun-owner/services/golang-forever.yaml",
			Contents: `description: golang is the best
repositories:
    - golang-forever/implementation
alertTarget: someone@example.com
developmentOnly: false
operationType: WORKLOAD
lifecycle: experimental`,
		},
	}

	docs.When("When BitBucket sends a webhook with valid payload")
	body := bitbucketserver.PullRequestOpenedPayload{
		Date:     bitbucketserver.Date(time.Now()),
		EventKey: bitbucketserver.PullRequestOpenedEvent,
		Actor:    bitbucketserver.User{
			// don't care
		},
		PullRequest: bitbucketserver.PullRequest{
			ID:          42,
			Title:       "some pr title",
			Description: "some pr description",
			FromRef: bitbucketserver.RepositoryReference{
				ID: "e2d2000000000000000000000000000000000000", // pr head
			},
			ToRef: bitbucketserver.RepositoryReference{
				ID: "e100000000000000000000000000000000000000", // mainline
			},
			Locked: false,
			Author: bitbucketserver.PullRequestParticipant{},
		},
	}
	bodyBytes, err := json.Marshal(&body)
	require.Nil(t, err)
	request, err := http.NewRequest(http.MethodPost, ts.URL+"/webhook/bitbucket", bytes.NewReader(bodyBytes))
	require.Nil(t, err)
	request.Header.Set("X-Event-Key", string(bitbucketserver.PullRequestOpenedEvent))
	rawResponse, err := http.DefaultClient.Do(request)
	require.Nil(t, err)
	response, err := tstWebResponseFromResponse(rawResponse)
	require.Nil(t, err)

	docs.Then("Then the request is successful")
	tstAssertNoBody(t, response, err, http.StatusNoContent)

	docs.Then("And the expected interactions with the BitBucket API have occurred")
	require.EqualValues(t, []string{
		"GetChangedFilesOnPullRequest(42)",
		"CreatePullRequestComment(42, all changed files are valid|)",
		"AddCommitBuildStatus(e2d2000000000000000000000000000000000000, metadata-service, true)",
	}, bbImpl.Recording)
}

func TestPOSTWebhookBitbucket_InvalidPR(t *testing.T) {
	tstReset()

	docs.Given("Given a pull request in BitBucket with invalid file edits")
	bbImpl.PRHead = "e2d2000000000000000000000000000000000000"
	bbImpl.ChangedFilesResponse = []repository.File{
		{
			Path:     "owners/fun-owner/repositories/golang-forever.implementation.yaml",
			Contents: `unknown: field`,
		},
	}

	docs.When("When BitBucket sends a webhook with valid payload")
	body := bitbucketserver.PullRequestOpenedPayload{
		Date:     bitbucketserver.Date(time.Now()),
		EventKey: bitbucketserver.PullRequestOpenedEvent,
		Actor:    bitbucketserver.User{
			// don't care
		},
		PullRequest: bitbucketserver.PullRequest{
			ID:          42,
			Title:       "some pr title",
			Description: "some pr description",
			FromRef: bitbucketserver.RepositoryReference{
				ID: "e2d2000000000000000000000000000000000000", // pr head
			},
			ToRef: bitbucketserver.RepositoryReference{
				ID: "e100000000000000000000000000000000000000", // mainline
			},
			Locked: false,
			Author: bitbucketserver.PullRequestParticipant{},
		},
	}
	bodyBytes, err := json.Marshal(&body)
	require.Nil(t, err)
	request, err := http.NewRequest(http.MethodPost, ts.URL+"/webhook/bitbucket", bytes.NewReader(bodyBytes))
	require.Nil(t, err)
	request.Header.Set("X-Event-Key", string(bitbucketserver.PullRequestOpenedEvent))
	rawResponse, err := http.DefaultClient.Do(request)
	require.Nil(t, err)
	response, err := tstWebResponseFromResponse(rawResponse)
	require.Nil(t, err)

	docs.Then("Then the request is successful")
	tstAssertNoBody(t, response, err, http.StatusNoContent)

	docs.Then("And the expected interactions with the BitBucket API have occurred")
	require.EqualValues(t, []string{
		"GetChangedFilesOnPullRequest(42)",
		"CreatePullRequestComment(42, # yaml validation failure||There were validation errors in changed files. Please fix yaml syntax and/or remove unknown fields:|| - failed to parse `owners/fun-owner/repositories/golang-forever.implementation.yaml`:|   yaml: unmarshal errors:|     line 1: field unknown not found in type openapi.RepositoryDto|)",
		"AddCommitBuildStatus(e2d2000000000000000000000000000000000000, metadata-service, false)",
	}, bbImpl.Recording)
}

func TestPOSTWebhookBitbucket_InvalidPayload(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they send a webhook with invalid payload")
	body := bitbucketserver.PullRequestOpenedPayload{} // hopefully invalid
	response, err := tstPerformPost("/webhook/bitbucket", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "webhook-invalid.json")
}
