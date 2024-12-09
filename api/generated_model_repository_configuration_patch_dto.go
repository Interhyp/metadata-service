/*
Metadata

Obtain and manage metadata for owners, services, repositories. Please see [README](https://github.com/Interhyp/metadata-service/blob/main/README.md) for details. **CLIENTS MUST READ!**

API version: v1
Contact: somebody@some-organisation.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

// RepositoryConfigurationPatchDto Attributes to configure the repository. If a configuration exists there are also some configured defaults for the repository.
type RepositoryConfigurationPatchDto struct {
	// Ssh-Keys configured on the repository.
	AccessKeys  []RepositoryConfigurationAccessKeyDto  `yaml:"accessKeys,omitempty" json:"accessKeys,omitempty"`
	MergeConfig *RepositoryConfigurationDtoMergeConfig `yaml:"mergeConfig,omitempty" json:"mergeConfig,omitempty"`
	// Use an explicit branch name regex.
	BranchNameRegex *string `yaml:"branchNameRegex,omitempty" json:"branchNameRegex,omitempty"`
	// Use an explicit commit message regex.
	CommitMessageRegex *string `yaml:"commitMessageRegex,omitempty" json:"commitMessageRegex,omitempty"`
	// Adds a corresponding commit message regex.
	CommitMessageType *string `yaml:"commitMessageType,omitempty" json:"commitMessageType,omitempty"`
	// Set the required successful builds counter.
	RequireSuccessfulBuilds *int32 `yaml:"requireSuccessfulBuilds,omitempty" json:"requireSuccessfulBuilds,omitempty"`
	// Set the required approvals counter.
	RequireApprovals *int32 `yaml:"requireApprovals,omitempty" json:"requireApprovals,omitempty"`
	// Exclude merge commits from commit checks.
	ExcludeMergeCommits *bool `yaml:"excludeMergeCommits,omitempty" json:"excludeMergeCommits,omitempty"`
	// Exclude users from commit checks.
	ExcludeMergeCheckUsers []ExcludeMergeCheckUserDto          `yaml:"excludeMergeCheckUsers,omitempty" json:"excludeMergeCheckUsers,omitempty"`
	Webhooks               *RepositoryConfigurationWebhooksDto `yaml:"webhooks,omitempty" json:"webhooks,omitempty"`
	// Map of string (group name e.g. some-owner) of strings (list of approvers), one approval for each group is required.
	Approvers map[string][]string `yaml:"approvers,omitempty" json:"approvers,omitempty"`
	// List of strings (list of watchers, either usernames or group identifier), which are added as reviewers but require no approval.
	Watchers []string `yaml:"watchers,omitempty" json:"watchers,omitempty"`
	// Moves the repository into the archive.
	Archived *bool `yaml:"archived,omitempty" json:"archived,omitempty"`
	// Repository will not be configured, also not archived.
	Unmanaged *bool `yaml:"unmanaged,omitempty" json:"unmanaged,omitempty"`
	// Control how the repository is used by GitHub Actions workflows in other repositories
	ActionsAccess *string `yaml:"actionsAccess,omitempty" json:"actionsAccess,omitempty"`
	// Configuration of conditional builds as map of structs (key name e.g. some-key) of target references.
	RequireConditions map[string]ConditionReferenceDto `yaml:"requireConditions,omitempty" json:"requireConditions,omitempty"`
}
