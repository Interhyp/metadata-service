/*
Metadata

Obtain and manage metadata for owners, services, repositories. Please see [README](https://github.com/Interhyp/metadata-service/blob/main/README.md) for details. **CLIENTS MUST READ!**

API version: v1
Contact: somebody@some-organisation.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

// RepositoryCreateDto struct for RepositoryCreateDto
type RepositoryCreateDto struct {
	// The alias of the repository owner
	Owner    string `yaml:"-" json:"owner"`
	Url      string `yaml:"url" json:"url"`
	Mainline string `yaml:"mainline" json:"mainline"`
	// the generator used for the initial contents of this repository
	Generator *string `yaml:"generator,omitempty" json:"generator,omitempty"`
	// this repository contains unit tests (currently ignored except for helm charts)
	Unittest      *bool                       `yaml:"unittest,omitempty" json:"unittest,omitempty"`
	Configuration *RepositoryConfigurationDto `yaml:"configuration,omitempty" json:"configuration,omitempty"`
	// Assign a category to a list of files, e.g. to mark them for caching purposes. The key is the category name, and the value is a list of paths. Files are considered to have that category if their path is in the list.
	Filecategory *map[string][]string `yaml:"filecategory,omitempty" json:"filecategory,omitempty"`
	// The jira issue to use for committing a change, or the last jira issue used.
	JiraIssue string `yaml:"-" json:"jiraIssue"`
	// A map of arbitrary string labels attached to this repository.
	Labels *map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}
