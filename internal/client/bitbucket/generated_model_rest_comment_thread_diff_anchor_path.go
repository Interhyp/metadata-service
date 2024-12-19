/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// RestCommentThreadDiffAnchorPath struct for RestCommentThreadDiffAnchorPath
type RestCommentThreadDiffAnchorPath struct {
    Extension *string `yaml:"extension,omitempty" json:"extension,omitempty"`
    Name *string `yaml:"name,omitempty" json:"name,omitempty"`
    Parent *string `yaml:"parent,omitempty" json:"parent,omitempty"`
    Components []string `yaml:"components,omitempty" json:"components,omitempty"`
}