/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// RestDeploymentSetRequest struct for RestDeploymentSetRequest
type RestDeploymentSetRequest struct {
    DeploymentSequenceNumber int64 `yaml:"deploymentSequenceNumber" json:"deploymentSequenceNumber"`
    Description string `yaml:"description" json:"description"`
    DisplayName string `yaml:"displayName" json:"displayName"`
    Environment RestDeploymentEnvironment `yaml:"environment" json:"environment"`
    Key string `yaml:"key" json:"key"`
    LastUpdated *int64 `yaml:"lastUpdated,omitempty" json:"lastUpdated,omitempty"`
    State string `yaml:"state" json:"state"`
    Url string `yaml:"url" json:"url"`
}
