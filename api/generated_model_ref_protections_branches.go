/*
Metadata

Obtain and manage metadata for owners, services, repositories. Please see [README](https://github.com/Interhyp/metadata-service/blob/main/README.md) for details. **CLIENTS MUST READ!**

API version: v1
Contact: somebody@some-organisation.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

// RefProtectionsBranches struct for RefProtectionsBranches
type RefProtectionsBranches struct {
	// Forces creating a PR to update the protected refs.
	RequirePR []ProtectedRef `yaml:"requirePR,omitempty" json:"requirePR,omitempty"`
	// Prevents all changes of the protected refs.
	PreventAllChanges []ProtectedRef `yaml:"preventAllChanges,omitempty" json:"preventAllChanges,omitempty"`
	// Prevents creation of the protected refs. Results in preventAllChanges for BitBucket.
	PreventCreation []ProtectedRef `yaml:"preventCreation,omitempty" json:"preventCreation,omitempty"`
	// Prevents deletion of the protected refs.
	PreventDeletion []ProtectedRef `yaml:"preventDeletion,omitempty" json:"preventDeletion,omitempty"`
	// Prevents pushes to the protected refs. Results in preventAllChanges for BitBucket.
	PreventPush []ProtectedRef `yaml:"preventPush,omitempty" json:"preventPush,omitempty"`
	// Prevents force pushes to the protected refs for users with push permission.
	PreventForcePush []ProtectedRef `yaml:"preventForcePush,omitempty" json:"preventForcePush,omitempty"`
}
