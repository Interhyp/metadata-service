/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// RestRestrictionRequest struct for RestRestrictionRequest
type RestRestrictionRequest struct {
    AccessKeyIds []int32 `yaml:"accessKeyIds,omitempty" json:"accessKeyIds,omitempty"`
    AccessKeys []RestSshAccessKey `yaml:"accessKeys,omitempty" json:"accessKeys,omitempty"`
    GroupNames []string `yaml:"groupNames,omitempty" json:"groupNames,omitempty"`
    Groups []string `yaml:"groups,omitempty" json:"groups,omitempty"`
    Id *int32 `yaml:"id,omitempty" json:"id,omitempty"`
    Matcher *UpdatePullRequestCondition1RequestSourceMatcher `yaml:"matcher,omitempty" json:"matcher,omitempty"`
    Scope *RestRestrictionRequestScope `yaml:"scope,omitempty" json:"scope,omitempty"`
    Type *string `yaml:"type,omitempty" json:"type,omitempty"`
    UserSlugs []string `yaml:"userSlugs,omitempty" json:"userSlugs,omitempty"`
    Users []RestApplicationUser `yaml:"users,omitempty" json:"users,omitempty"`
}
