/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// RestBuildStatus struct for RestBuildStatus
type RestBuildStatus struct {
    Name *string `yaml:"name,omitempty" json:"name,omitempty"`
    Key *string `yaml:"key,omitempty" json:"key,omitempty"`
    Parent *string `yaml:"parent,omitempty" json:"parent,omitempty"`
    State *string `yaml:"state,omitempty" json:"state,omitempty"`
    Ref *string `yaml:"ref,omitempty" json:"ref,omitempty"`
    Duration *int64 `yaml:"duration,omitempty" json:"duration,omitempty"`
    TestResults *RestBuildStatusTestResults `yaml:"testResults,omitempty" json:"testResults,omitempty"`
    CreatedDate *int64 `yaml:"createdDate,omitempty" json:"createdDate,omitempty"`
    UpdatedDate *int64 `yaml:"updatedDate,omitempty" json:"updatedDate,omitempty"`
    Description *string `yaml:"description,omitempty" json:"description,omitempty"`
    BuildNumber *string `yaml:"buildNumber,omitempty" json:"buildNumber,omitempty"`
    Url *string `yaml:"url,omitempty" json:"url,omitempty"`
}