/*
Bitbucket Data Center

This is the reference document for the Atlassian Bitbucket REST API. The REST API is for developers who want to:    - integrate Bitbucket with other applications;   - create scripts that interact with Bitbucket; or   - develop plugins that enhance the Bitbucket UI, using REST to interact with the backend.    You can read more about developing Bitbucket plugins in the [Bitbucket Developer Documentation](https://developer.atlassian.com/bitbucket/server/docs/latest/).

API version: 8.19
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package bitbucketclient


// RestInsightReport struct for RestInsightReport
type RestInsightReport struct {
    Result *string `yaml:"result,omitempty" json:"result,omitempty"`
    Key *string `yaml:"key,omitempty" json:"key,omitempty"`
    CreatedDate *float32 `yaml:"createdDate,omitempty" json:"createdDate,omitempty"`
    Reporter *string `yaml:"reporter,omitempty" json:"reporter,omitempty"`
    Data []RestInsightReportData `yaml:"data,omitempty" json:"data,omitempty"`
    Title *string `yaml:"title,omitempty" json:"title,omitempty"`
    Details *string `yaml:"details,omitempty" json:"details,omitempty"`
    Link *string `yaml:"link,omitempty" json:"link,omitempty"`
    LogoUrl *string `yaml:"logoUrl,omitempty" json:"logoUrl,omitempty"`
}