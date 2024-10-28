package acceptance

import (
	"bytes"
	"encoding/json"
	"github.com/Interhyp/go-backend-service-common/docs"
	bitbucketserver "github.com/go-playground/webhooks/v6/bitbucket-server"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestPOSTWebhookBitbucket_Success(t *testing.T) {
	tstReset()

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
	request, err := http.NewRequest(http.MethodPost, ts.URL+"/webhooks/vcs/bitbucket_datacenter", bytes.NewReader(bodyBytes))
	require.Nil(t, err)
	request.Header.Set("X-Event-Key", string(bitbucketserver.PullRequestOpenedEvent))
	rawResponse, err := http.DefaultClient.Do(request)
	require.Nil(t, err)
	response, err := tstWebResponseFromResponse(rawResponse)
	require.Nil(t, err)

	docs.Then("Then the request is successful")
	tstAssertNoBody(t, response, err, http.StatusNoContent)
}

func TestPOSTWebhookBitbucket_InvalidPR(t *testing.T) {
	tstReset()

	docs.When("When BitBucket sends a webhook with valid payload")
	body := bitbucketserver.PullRequestOpenedPayload{
		Date:     bitbucketserver.Date(time.Now()),
		EventKey: bitbucketserver.PullRequestOpenedEvent,
		Actor:    bitbucketserver.User{
			// don't care
		},
		PullRequest: bitbucketserver.PullRequest{
			ID:          43,
			Title:       "some pr title",
			Description: "some pr description",
			FromRef: bitbucketserver.RepositoryReference{
				ID: "e2d3000000000000000000000000000000000000", // pr head
			},
			ToRef: bitbucketserver.RepositoryReference{
				ID: "e200000000000000000000000000000000000000", // mainline
			},
			Locked: false,
			Author: bitbucketserver.PullRequestParticipant{},
		},
	}
	bodyBytes, err := json.Marshal(&body)
	require.Nil(t, err)
	request, err := http.NewRequest(http.MethodPost, ts.URL+"/webhooks/vcs/bitbucket_datacenter", bytes.NewReader(bodyBytes))
	require.Nil(t, err)
	request.Header.Set("X-Event-Key", string(bitbucketserver.PullRequestOpenedEvent))
	rawResponse, err := http.DefaultClient.Do(request)
	require.Nil(t, err)
	response, err := tstWebResponseFromResponse(rawResponse)
	require.Nil(t, err)

	docs.Then("Then the request is successful")
	tstAssertNoBody(t, response, err, http.StatusNoContent)
}

func TestPOSTWebhookBitbucket_InvalidPayload(t *testing.T) {
	tstReset()

	docs.Given("Given an unauthenticated user")
	token := tstUnauthenticated()

	docs.When("When they send a webhook with invalid payload")
	body := bitbucketserver.PullRequestOpenedPayload{} // hopefully invalid
	response, err := tstPerformPost("/webhooks/vcs/bitbucket_datacenter", token, &body)

	docs.Then("Then the request fails and the error response is as expected")
	tstAssert(t, response, err, http.StatusBadRequest, "webhook-invalid.json")
}
