/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// RestDiffHunk struct for RestDiffHunk
type RestDiffHunk struct {
    Context *string `yaml:"context,omitempty" json:"context,omitempty"`
    SourceLine *int32 `yaml:"sourceLine,omitempty" json:"sourceLine,omitempty"`
    Segments []RestDiffSegment `yaml:"segments,omitempty" json:"segments,omitempty"`
    SourceSpan *int32 `yaml:"sourceSpan,omitempty" json:"sourceSpan,omitempty"`
    DestinationSpan *int32 `yaml:"destinationSpan,omitempty" json:"destinationSpan,omitempty"`
    DestinationLine *int32 `yaml:"destinationLine,omitempty" json:"destinationLine,omitempty"`
    Truncated *bool `yaml:"truncated,omitempty" json:"truncated,omitempty"`
}
