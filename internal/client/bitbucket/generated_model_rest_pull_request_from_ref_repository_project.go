/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// RestPullRequestFromRefRepositoryProject struct for RestPullRequestFromRefRepositoryProject
type RestPullRequestFromRefRepositoryProject struct {
    Name *string `yaml:"name,omitempty" json:"name,omitempty" validate:"regexp=^[^~].*"`
    Key string `yaml:"key" json:"key"`
    Public *bool `yaml:"public,omitempty" json:"public,omitempty"`
    Id *int32 `yaml:"id,omitempty" json:"id,omitempty"`
    Type *string `yaml:"type,omitempty" json:"type,omitempty"`
    AvatarUrl *string `yaml:"-" json:"-"`
    Description *string `yaml:"description,omitempty" json:"description,omitempty"`
        // Deprecated
    Namespace *string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
    Scope *string `yaml:"scope,omitempty" json:"scope,omitempty"`
    Avatar *string `yaml:"avatar,omitempty" json:"avatar,omitempty"`
    Links map[string]interface{} `yaml:"-" json:"-"`
}
