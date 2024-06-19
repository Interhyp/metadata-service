/*
Metadata

Obtain and manage metadata for owners, services, repositories. Please see [README](https://github.com/Interhyp/metadata-service/blob/main/README.md) for details. **CLIENTS MUST READ!**

API version: v1
Contact: somebody@some-organisation.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

// RepositoryListDto struct for RepositoryListDto
type RepositoryListDto struct {
	Repositories map[string]RepositoryDto `yaml:"repositories" json:"repositories"`
	// ISO-8601 UTC date time at which the list of repositories was obtained from service-metadata
	TimeStamp string `yaml:"-" json:"timeStamp"`
}
