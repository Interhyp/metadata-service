/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// RestAutoMergeRequest struct for RestAutoMergeRequest
type RestAutoMergeRequest struct {
    Message *string `yaml:"message,omitempty" json:"message,omitempty"`
    AutoSubject *bool `yaml:"autoSubject,omitempty" json:"autoSubject,omitempty"`
    StrategyId *string `yaml:"strategyId,omitempty" json:"strategyId,omitempty"`
    CreatedDate *int64 `yaml:"createdDate,omitempty" json:"createdDate,omitempty"`
    FromHash *string `yaml:"fromHash,omitempty" json:"fromHash,omitempty"`
    ToRefId *string `yaml:"toRefId,omitempty" json:"toRefId,omitempty"`
}
