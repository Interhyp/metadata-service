/*
Metadata

Obtain and manage metadata for owners, services, repositories. Please see [README](https://github.com/Interhyp/metadata-service/blob/main/README.md) for details. **CLIENTS MUST READ!**

API version: v1
Contact: somebody@some-organisation.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

// ConditionReferenceDto Configuration of conditional build references.
type ConditionReferenceDto struct {
	// Reference of a branch.
	RefMatcher string `yaml:"refMatcher" json:"refMatcher"`
	// list of groups, NPAs or app ids for whom this protection does not apply. Individual users are not supported.
	Exemptions []string `yaml:"exemptions,omitempty" json:"exemptions,omitempty"`
}
