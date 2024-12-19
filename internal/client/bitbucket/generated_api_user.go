/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient

import (
    "context"
    "fmt"
    aurestclientapi "github.com/StephanHCB/go-autumn-restclient/api"
    "net/http"
    urlUtil "net/url"
    "strings"
)

type UserAPI interface {

    /*
       GetUser Get user
    */
    GetUser(ctx context.Context, userSlug string) (RestApplicationUser, aurestclientapi.ParsedResponse, error)

    // GetUserExecutes the request
    // @return RestApplicationUser
    GetUserRequest(ctx context.Context, userSlug string) UserAPIGetUserRequest
}

 type UserAPIGetUserRequest struct {
    ctx context.Context
    ApiService *UserAPIRepository
    userSlug string
}

func (r *UserAPIGetUserRequest) Execute() (RestApplicationUser, aurestclientapi.ParsedResponse, error) {
    return r.ApiService.GetUserExecute(r)
}

func (a *UserAPIRepository) GetUserRequest(ctx context.Context, userSlug string) UserAPIGetUserRequest {
    return UserAPIGetUserRequest{
        ApiService: a,
        ctx: ctx,
        userSlug: userSlug,
    }
}

func (a *UserAPIRepository) GetUserExecute(r *UserAPIGetUserRequest) (RestApplicationUser, aurestclientapi.ParsedResponse, error) {
    fullUrlValue := a.baseUrl() + "/api/latest/users/{userSlug}"
    fullUrlValue = strings.ReplaceAll(fullUrlValue, "{userSlug}", urlUtil.PathEscape(r.userSlug))
    requestURL, _ := urlUtil.Parse(fullUrlValue)
    return a.makeGetUserCall(r.ctx, requestURL, nil)
}

func (a *UserAPIRepository) GetUser(ctx context.Context, userSlug string) (RestApplicationUser, aurestclientapi.ParsedResponse, error) {
    fullUrlValue := a.baseUrl() + "/api/latest/users/{userSlug}"
    fullUrlValue = strings.ReplaceAll(fullUrlValue, "{userSlug}", urlUtil.PathEscape(userSlug))
    requestURL, _ := urlUtil.Parse(fullUrlValue)
    return a.makeGetUserCall(ctx, requestURL, nil)
}

func (a *UserAPIRepository) makeGetUserCall(ctx context.Context, requestURL *urlUtil.URL, requestBody any) (RestApplicationUser, aurestclientapi.ParsedResponse, error) {
	method := http.MethodGet
	requestUrl := requestURL.String()

    var result RestApplicationUser
    emptyResponse := make([]byte, 0)
    responseBodyPointer := &emptyResponse
    response := aurestclientapi.ParsedResponse{
        Body: &responseBodyPointer,
    }
    err := a.httpClient().Perform(ctx, method, requestUrl, requestBody, &response)
	if err != nil {
		return result,response, err
	}
    if response.Status == 401 {
		err = safeUnmarshal[DismissRetentionConfigReviewNotification401Response](&response)
        if err == nil {
            err = NewError(fmt.Sprintf("Got status %d", response.Status), response.Status)
        }
        return result, response, err
    }
    if response.Status == 404 {
		err = safeUnmarshal[DismissRetentionConfigReviewNotification401Response](&response)
        if err == nil {
            err = NewError(fmt.Sprintf("Got status %d", response.Status), response.Status)
        }
        return result, response, err
    }

    if response.Status < 400 {
        err = safeUnmarshal[RestApplicationUser](&response)
        if err == nil {
            result = response.Body.(RestApplicationUser)
        }
    } else {
        err = NewError(fmt.Sprintf("Got unknown status %d", response.Status), response.Status)
    }
    return result, response, err
}

type UserAPIRepository struct {
    ApiClient *ApiClient
}

func (c *UserAPIRepository) baseUrl() string {
    return c.ApiClient.BaseUrl
}

func (c *UserAPIRepository) httpClient() aurestclientapi.Client {
    return c.ApiClient.Client
}

func NewUserAPI(client *ApiClient) UserAPI {
    return &UserAPIRepository{ApiClient: client}
}