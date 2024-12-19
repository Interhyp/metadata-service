/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// RestErrorMessage struct for RestErrorMessage
type RestErrorMessage struct {
    Message *string `yaml:"message,omitempty" json:"message,omitempty"`
    Context *string `yaml:"context,omitempty" json:"context,omitempty"`
    ExceptionName *string `yaml:"exceptionName,omitempty" json:"exceptionName,omitempty"`
}