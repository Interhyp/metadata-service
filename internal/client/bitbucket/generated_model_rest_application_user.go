/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// RestApplicationUser struct for RestApplicationUser
type RestApplicationUser struct {
    Name *string `yaml:"name,omitempty" json:"name,omitempty"`
    Id *int32 `yaml:"id,omitempty" json:"id,omitempty"`
    Type *string `yaml:"type,omitempty" json:"type,omitempty"`
    Active *bool `yaml:"active,omitempty" json:"active,omitempty"`
    DisplayName *string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
    EmailAddress *string `yaml:"emailAddress,omitempty" json:"emailAddress,omitempty"`
    Links map[string]interface{} `yaml:"-" json:"-"`
    Slug *string `yaml:"slug,omitempty" json:"slug,omitempty"`
    AvatarUrl *string `yaml:"-" json:"-"`
}
