/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// RestGitTagCreateRequest struct for RestGitTagCreateRequest
type RestGitTagCreateRequest struct {
    Force *bool `yaml:"force,omitempty" json:"force,omitempty"`
    Message *string `yaml:"message,omitempty" json:"message,omitempty"`
    Name *string `yaml:"name,omitempty" json:"name,omitempty"`
    StartPoint *string `yaml:"startPoint,omitempty" json:"startPoint,omitempty"`
    Type *string `yaml:"type,omitempty" json:"type,omitempty"`
}
