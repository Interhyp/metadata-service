/*
Metadata

Obtain and manage metadata for owners, services, repositories. Please see [README](https://github.com/Interhyp/metadata-service/blob/main/README.md) for details. **CLIENTS MUST READ!**

API version: v1
Contact: somebody@some-organisation.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

// DeletionDto struct for DeletionDto
type DeletionDto struct {
	// The jira issue to use for committing the deletion.
	JiraIssue string `yaml:"-" json:"jiraIssue"`
}
