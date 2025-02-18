/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// GetCommits200Response struct for GetCommits200Response
type GetCommits200Response struct {
    Values []RestCommit `yaml:"values,omitempty" json:"values,omitempty"`
    Size *float32 `yaml:"size,omitempty" json:"size,omitempty"`
    Limit *float32 `yaml:"limit,omitempty" json:"limit,omitempty"`
    NextPageStart *int32 `yaml:"nextPageStart,omitempty" json:"nextPageStart,omitempty"`
    IsLastPage *bool `yaml:"isLastPage,omitempty" json:"isLastPage,omitempty"`
    Start *int32 `yaml:"start,omitempty" json:"start,omitempty"`
}
