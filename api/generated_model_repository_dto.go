/*
Metadata

Obtain and manage metadata for owners, services, repositories. Please see [README](https://github.com/Interhyp/metadata-service/blob/main/README.md) for details. **CLIENTS MUST READ!**

API version: v1
Contact: somebody@some-organisation.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

// RepositoryDto struct for RepositoryDto
type RepositoryDto struct {
	// The type of the repository as determined by its key.
	Type *string `yaml:"-" json:"type,omitempty"`
	// The alias of the repository owner
	Owner       string  `yaml:"-" json:"owner"`
	Description *string `yaml:"description,omitempty" json:"description,omitempty"`
	Url         string  `yaml:"url" json:"url"`
	Mainline    string  `yaml:"mainline" json:"mainline"`
	// the generator used for the initial contents of this repository
	Generator     *string                     `yaml:"generator,omitempty" json:"generator,omitempty"`
	Configuration *RepositoryConfigurationDto `yaml:"configuration,omitempty" json:"configuration,omitempty"`
	// ISO-8601 UTC date time at which this information was originally committed. When sending an update, include the original timestamp you got so we can detect concurrent updates.
	TimeStamp string `yaml:"-" json:"timeStamp"`
	// The git commit hash this information was originally committed under. When sending an update, include the original commitHash you got so we can detect concurrent updates.
	CommitHash string `yaml:"-" json:"commitHash"`
	// The jira issue to use for committing a change, or the last jira issue used.
	JiraIssue string `yaml:"-" json:"jiraIssue"`
	// A map of arbitrary string labels attached to this repository.
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}
