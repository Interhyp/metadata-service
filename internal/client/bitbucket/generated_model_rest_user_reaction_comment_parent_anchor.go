/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// RestUserReactionCommentParentAnchor struct for RestUserReactionCommentParentAnchor
type RestUserReactionCommentParentAnchor struct {
    Path *RestCommentThreadDiffAnchorPath `yaml:"path,omitempty" json:"path,omitempty"`
    DiffType *string `yaml:"diffType,omitempty" json:"diffType,omitempty"`
    FileType *string `yaml:"fileType,omitempty" json:"fileType,omitempty"`
    FromHash *string `yaml:"fromHash,omitempty" json:"fromHash,omitempty"`
    LineType *string `yaml:"lineType,omitempty" json:"lineType,omitempty"`
    PullRequest *RestCommentThreadDiffAnchorPullRequest `yaml:"pullRequest,omitempty" json:"pullRequest,omitempty"`
    LineComment *bool `yaml:"lineComment,omitempty" json:"lineComment,omitempty"`
    Line *int32 `yaml:"line,omitempty" json:"line,omitempty"`
    SrcPath *RestCommentThreadDiffAnchorPath `yaml:"srcPath,omitempty" json:"srcPath,omitempty"`
    ToHash *string `yaml:"toHash,omitempty" json:"toHash,omitempty"`
}