/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// RestRequiredBuildConditionSetRequest struct for RestRequiredBuildConditionSetRequest
type RestRequiredBuildConditionSetRequest struct {
        // A non-empty list of build parent keys that require green builds for this merge check to pass
    BuildParentKeys []string `yaml:"buildParentKeys" json:"buildParentKeys"`
    ExemptRefMatcher *RestRefMatcher `yaml:"exemptRefMatcher,omitempty" json:"exemptRefMatcher,omitempty"`
    RefMatcher UpdatePullRequestCondition1RequestSourceMatcher `yaml:"refMatcher" json:"refMatcher"`
}
