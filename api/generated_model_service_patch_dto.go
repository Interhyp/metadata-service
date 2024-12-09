/*
Metadata

Obtain and manage metadata for owners, services, repositories. Please see [README](https://github.com/Interhyp/metadata-service/blob/main/README.md) for details. **CLIENTS MUST READ!**

API version: v1
Contact: somebody@some-organisation.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

// ServicePatchDto struct for ServicePatchDto
type ServicePatchDto struct {
	// The alias of the service owner. Note, a patch with changed owner will move the service and any associated repositories to the new owner, but of course this will not move e.g. Jenkins jobs. That's your job.
	Owner *string `yaml:"owner,omitempty" json:"owner,omitempty"`
	// A short description of the functionality of the service.
	Description *string `yaml:"description,omitempty" json:"description,omitempty"`
	// A list of quicklinks related to the service
	Quicklinks []Quicklink `yaml:"quicklinks,omitempty" json:"quicklinks,omitempty"`
	// The keys of repositories associated with the service. When sending an update, they must refer to repositories that belong to this service, or the update will fail
	Repositories []string `yaml:"repositories,omitempty" json:"repositories,omitempty"`
	// The default channel used to send any alerts of the service to. Can be an email address or a Teams webhook URL
	AlertTarget *string `yaml:"alertTarget,omitempty" json:"alertTarget,omitempty"`
	// The operation type of the service. 'WORKLOAD' follows the default deployment strategy of one instance per environment, 'PLATFORM' one instance per cluster or node and 'APPLICATION' is a standalone application that is not deployed via the common strategies.
	OperationType *string `yaml:"operationType,omitempty" json:"operationType,omitempty"`
	// The value defines if the service is available from the internet and the time period in which security holes must be processed.
	InternetExposed *bool             `yaml:"internetExposed,omitempty" json:"internetExposed,omitempty"`
	Tags            []string          `yaml:"tags,omitempty" json:"tags,omitempty"`
	Labels          map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Spec            *ServiceSpecDto   `yaml:"spec,omitempty" json:"spec,omitempty"`
	// Post promote dependencies.
	PostPromotes *PostPromote `yaml:"postPromotes,omitempty" json:"postPromotes,omitempty"`
	// ISO-8601 UTC date time at which this information was originally committed. When sending an update, include the original timestamp you got so we can detect concurrent updates.
	TimeStamp string `yaml:"-" json:"timeStamp"`
	// The git commit hash this information was originally committed under. When sending an update, include the original commitHash you got so we can detect concurrent updates.
	CommitHash string `yaml:"-" json:"commitHash"`
	// The jira issue to use for committing a change, or the last jira issue used.
	JiraIssue string `yaml:"-" json:"jiraIssue"`
	// The current phase of the service's development. A service usually starts off as 'experimental', then becomes 'operational' (i. e. can be reliably used and/or consumed). Once 'deprecated', the service doesn’t guarantee reliable use/consumption any longer.
	Lifecycle *string `yaml:"lifecycle,omitempty" json:"lifecycle,omitempty"`
}
