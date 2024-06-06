package bbclient

import (
	"context"
	"fmt"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/errors/httperror"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/repository/bitbucket/bbclientint"
	auapmclient "github.com/StephanHCB/go-autumn-restclient-apm/implementation/client"
	aurestbreakerprometheus "github.com/StephanHCB/go-autumn-restclient-circuitbreaker-prometheus"
	aurestbreaker "github.com/StephanHCB/go-autumn-restclient-circuitbreaker/implementation/breaker"
	aurestclientprometheus "github.com/StephanHCB/go-autumn-restclient-prometheus"
	aurestclientapi "github.com/StephanHCB/go-autumn-restclient/api"
	aurestcaching "github.com/StephanHCB/go-autumn-restclient/implementation/caching"
	auresthttpclient "github.com/StephanHCB/go-autumn-restclient/implementation/httpclient"
	aurestrecorder "github.com/StephanHCB/go-autumn-restclient/implementation/recorder"
	aurestlogging "github.com/StephanHCB/go-autumn-restclient/implementation/requestlogging"
	aurestretry "github.com/StephanHCB/go-autumn-restclient/implementation/retry"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/go-http-utils/headers"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Vault               librepo.Vault

	apiBaseUrl string

	Client        aurestclientapi.Client
	NoRetryClient aurestclientapi.Client
}

func New(configuration librepo.Configuration, logging librepo.Logging, vault librepo.Vault) bbclientint.BitbucketClient {
	return &Impl{
		Configuration:       configuration,
		CustomConfiguration: config.Custom(configuration),
		Logging:             logging,
		Vault:               vault,
	}
}

// --- setup ---

func (c *Impl) Setup() error {
	c.apiBaseUrl = c.CustomConfiguration.BitbucketServer() + "/bitbucket"

	client, err := auresthttpclient.New(0, nil, c.requestHeaderManipulator())
	if err != nil {
		return err
	}
	aurestclientprometheus.InstrumentHttpClient(client)

	logWrapper := aurestlogging.New(client)

	circuitBreakerWrapper := aurestbreaker.New(
		logWrapper,
		"bitbucket",
		100,
		5*time.Minute,
		60*time.Second,
		// includes possible retries, once the context is cancelled further requests will fail directly
		15*time.Second,
	)
	aurestbreakerprometheus.InstrumentCircuitBreakerClient(circuitBreakerWrapper)

	// allow tests to pre-populate
	if c.NoRetryClient == nil {
		c.NoRetryClient = circuitBreakerWrapper
	}

	apmClientWrapper := auapmclient.New(circuitBreakerWrapper)

	retryWrapper := aurestretry.New(
		apmClientWrapper,
		3,
		c.retryCondition(),
		c.betweenFailureAndRetry(),
	)
	aurestclientprometheus.InstrumentRetryClient(retryWrapper)

	recordingWrapper := aurestrecorder.New(retryWrapper)

	cacheWrapper := aurestcaching.New(
		recordingWrapper,
		func(ctx context.Context, method string, url string, requestBody interface{}) bool {
			return method == http.MethodGet && strings.Contains(url, fmt.Sprintf("%s/users", bbclientint.CoreApi))
		},
		func(ctx context.Context, method string, url string, requestBody interface{}, response *aurestclientapi.ParsedResponse) bool {
			return response != nil && response.Status == http.StatusOK && strings.Contains(url, fmt.Sprintf("%s/users", bbclientint.CoreApi))
		},
		nil,
		time.Duration(c.CustomConfiguration.BitbucketCacheRetentionSeconds())*time.Second,
		c.CustomConfiguration.BitbucketCacheSize(),
	)
	aurestclientprometheus.InstrumentCacheClient(cacheWrapper)

	//allow tests to pre-populate
	if c.Client == nil {
		c.Client = cacheWrapper
	}

	return nil
}

func (c *Impl) requestHeaderManipulator() func(ctx context.Context, r *http.Request) {
	return func(ctx context.Context, r *http.Request) {
		if r.Method != http.MethodPost {
			r.Header.Set(headers.Accept, aurestclientapi.ContentTypeApplicationJson)
		}
		if ctx.Value("authorization") != nil && ctx.Value("authorization") != "" && r.Method == http.MethodPost {
			r.Header.Set("Authorization", ctx.Value("authorization").(string))
		} else {
			r.SetBasicAuth(c.CustomConfiguration.BitbucketUsername(), c.CustomConfiguration.BitbucketPassword())
		}
	}
}

func (c *Impl) retryCondition() aurestclientapi.RetryConditionCallback {
	return func(_ context.Context, response *aurestclientapi.ParsedResponse, err error) bool {
		// bitbucket sometimes does this rather randomly, we just retry up to 3 times
		return response.Status == http.StatusServiceUnavailable
	}
}

func (c *Impl) betweenFailureAndRetry() aurestclientapi.BeforeRetryCallback {
	return func(ctx context.Context, originalResponse *aurestclientapi.ParsedResponse, originalError error) error {
		c.Logging.Logger().Ctx(ctx).Warn().Print("got 503 from bitbucket - retrying request")
		return nil
	}
}

// --- request implementations ---

func (c *Impl) call(ctx context.Context, method string, requestUrlExtension string, requestBody interface{}, responseBodyPointer interface{}) error {
	remoteUrl := fmt.Sprintf("%s/%s", c.apiBaseUrl, requestUrlExtension)
	response := &aurestclientapi.ParsedResponse{
		Body: responseBodyPointer,
	}
	err := c.Client.Perform(ctx, method, remoteUrl, requestBody, response)
	if err != nil {
		return err
	}

	switch response.Status {
	case
		http.StatusOK,
		http.StatusCreated,
		http.StatusNoContent:
		return nil
	}

	return httperror.New(ctx, fmt.Sprintf("received unexpected status %d from bitbucket %s %s", response.Status, method, requestUrlExtension), response.Status)
}

func (c *Impl) GetBitbucketUser(ctx context.Context, username string) (repository.BitbucketUser, error) {
	urlExt := fmt.Sprintf("%s/users/%s",
		bbclientint.CoreApi,
		url.PathEscape(username))
	response := repository.BitbucketUser{}
	err := c.call(ctx, http.MethodGet, urlExt, nil, &response)
	return response, err
}

func (c *Impl) GetPullRequest(ctx context.Context, projectKey string, repositorySlug string, pullRequestId int32) (bbclientint.PullRequest, error) {
	urlExt := fmt.Sprintf("%s/projects/%s/repos/%s/pull-requests/%d",
		bbclientint.CoreApi,
		url.PathEscape(projectKey),
		url.PathEscape(repositorySlug),
		pullRequestId)
	response := bbclientint.PullRequest{}
	err := c.call(ctx, http.MethodGet, urlExt, nil, &response)
	return response, err
}

func (c *Impl) GetChanges(ctx context.Context, projectKey string, repositorySlug string, sinceHash string, untilHash string) (bbclientint.Changes, error) {
	// since : main
	// until : pr head
	urlExt := fmt.Sprintf("%s/projects/%s/repos/%s/changes?since=%s&until=%s&limit=%d",
		bbclientint.CoreApi,
		url.PathEscape(projectKey),
		url.PathEscape(repositorySlug),
		url.QueryEscape(sinceHash),
		url.QueryEscape(untilHash),
		1000) // TODO pagination?
	response := bbclientint.Changes{}
	err := c.call(ctx, http.MethodGet, urlExt, nil, &response)
	return response, err
}

func (c *Impl) getFileContentsPage(ctx context.Context, projectKey string, repositorySlug string, atHash string, path string, start int, limit int) (bbclientint.PaginatedLines, error) {
	escapedPath := ""
	for _, pathComponent := range strings.Split(path, "/") {
		escapedPath += "/" + url.PathEscape(pathComponent)
	}
	urlExt := fmt.Sprintf("%s/projects/%s/repos/%s/browse/%s?at=%s&start=%d&limit=%d",
		bbclientint.CoreApi,
		url.PathEscape(projectKey),
		url.PathEscape(repositorySlug),
		escapedPath,
		url.QueryEscape(atHash),
		start,
		limit)
	response := bbclientint.PaginatedLines{}
	err := c.call(ctx, http.MethodGet, urlExt, nil, &response)
	return response, err
}

func (c *Impl) GetFileContentsAt(ctx context.Context, projectKey string, repositorySlug string, atHash string, path string) (string, error) {
	var contents strings.Builder
	var err error
	start := 0

	page := bbclientint.PaginatedLines{
		IsLastPage:    false,
		NextPageStart: &start,
	}

	for !page.IsLastPage && page.NextPageStart != nil {
		page, err = c.getFileContentsPage(ctx, projectKey, repositorySlug, atHash, path, *page.NextPageStart, 1000)
		if err != nil {
			return contents.String(), err
		}
		for _, line := range page.Lines {
			contents.WriteString(line.Text + "\n")
		}
	}

	return contents.String(), nil
}
