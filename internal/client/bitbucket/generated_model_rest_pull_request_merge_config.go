/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// RestPullRequestMergeConfig struct for RestPullRequestMergeConfig
type RestPullRequestMergeConfig struct {
    Type *string `yaml:"type,omitempty" json:"type,omitempty"`
    CommitMessageTemplate *RestPullRequestMergeConfigCommitMessageTemplate `yaml:"commitMessageTemplate,omitempty" json:"commitMessageTemplate,omitempty"`
    CommitSummaries *int32 `yaml:"commitSummaries,omitempty" json:"commitSummaries,omitempty"`
    DefaultStrategy *RestPullRequestMergeConfigDefaultStrategy `yaml:"defaultStrategy,omitempty" json:"defaultStrategy,omitempty"`
    Strategies []RestPullRequestMergeStrategy `yaml:"strategies,omitempty" json:"strategies,omitempty"`
}
