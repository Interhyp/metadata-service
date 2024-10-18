/*
Metadata

Obtain and manage metadata for owners, services, repositories. Please see [README](https://github.com/Interhyp/metadata-service/blob/main/README.md) for details. **CLIENTS MUST READ!**

API version: v1
Contact: somebody@some-organisation.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

// ProtectedRef struct for ProtectedRef
type ProtectedRef struct {
	// fnmatch pattern to define protected refs. Must not start with refs/heads/ or refs/tags/. Special value :MAINLINE: matches the currently configured mainline for branch protections.
	Pattern string `yaml:"pattern" json:"pattern" validate:"regexp=^(?!refs\\/(heads|tags)\\/).*$"`
	// list of users, teams or apps for whom this protection does not apply
	Exemptions []string `yaml:"exemptions,omitempty" json:"exemptions,omitempty"`
	// list of teams for whom this protection does not apply
	ExemptionsRoles []string `yaml:"exemptionsRoles,omitempty" json:"exemptionsRoles,omitempty"`
}
