package bitbucketclient

import (
	"context"
	"fmt"
	"github.com/Interhyp/go-backend-service-common/web/middleware/requestid"
	"github.com/Interhyp/metadata-service/pkg/recorder"
	auapmclient "github.com/StephanHCB/go-autumn-restclient-apm/implementation/client"
	aurestclientapi "github.com/StephanHCB/go-autumn-restclient/api"
	aurestrecorder "github.com/StephanHCB/go-autumn-restclient/implementation/recorder"
	"github.com/go-http-utils/headers"
	"net/http"
	urlUtil "net/url"
	"strings"
)

func NewClient(baseURL string, accessToken string) (*ApiClient, error) {
	clientConfig := DefaultApiClientConfig(fmt.Sprintf("%s/rest", baseURL))
	clientConfig.CachingConfigurer = nil
	clientConfig.RequestManipulator = func(ctx context.Context, request *http.Request) {
		request.Header.Set(headers.Accept, aurestclientapi.ContentTypeApplicationJson)

		request.Header.Set(headers.Authorization, fmt.Sprintf("Bearer %s", accessToken))

		if reqId := requestid.GetReqID(ctx); reqId != "" {
			request.Header.Set(requestid.RequestIDHeader, reqId)
		}
		auapmclient.AddTraceHeadersRequestManipulator(ctx, request)
	}
	clientConfig.RecorderConfigurer = func(client aurestclientapi.Client) aurestclientapi.Client {
		return aurestrecorder.New(client, aurestrecorder.RecorderOptions{
			ConstructFilenameFunc: recorder.ConstructFilenameV4,
		})
	}

	return NewApiClientConfigured(clientConfig)
}

func (r *RepositoryAPIGetContent1Request) FilePathCompatibleExecute() (GetContent1200Response, aurestclientapi.ParsedResponse, error) {
	escapedPath := ""
	for _, pathComponent := range strings.Split(r.path, "/") {
		escapedPath += "/" + urlUtil.PathEscape(pathComponent)
	}
	fullUrlValue := r.ApiService.baseUrl() + "/api/latest/projects/{projectKey}/repos/{repositorySlug}/browse{path}"
	fullUrlValue = strings.ReplaceAll(fullUrlValue, "{path}", escapedPath)
	fullUrlValue = strings.ReplaceAll(fullUrlValue, "{projectKey}", urlUtil.PathEscape(r.projectKey))
	fullUrlValue = strings.ReplaceAll(fullUrlValue, "{repositorySlug}", urlUtil.PathEscape(r.repositorySlug))
	requestURL, _ := urlUtil.Parse(fullUrlValue)
	if r.noContent != nil {
		withUrlQueryParam(requestURL, "noContent", *r.noContent)
	}
	if r.at != nil {
		withUrlQueryParam(requestURL, "at", *r.at)
	}
	if r.size != nil {
		withUrlQueryParam(requestURL, "size", *r.size)
	}
	if r.blame != nil {
		withUrlQueryParam(requestURL, "blame", *r.blame)
	}
	if r.type_ != nil {
		withUrlQueryParam(requestURL, "type", *r.type_)
	}
	return r.ApiService.makeGetContent1Call(r.ctx, requestURL, nil)
}
