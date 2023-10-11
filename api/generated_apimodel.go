// Code generated by interhyp-improved OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"time"
)

type _dummyTime struct {
	Timestamp *time.Time
}

type ConditionReferenceDto struct {
	// Reference of a branch.
	RefMatcher string `yaml:"refMatcher" json:"refMatcher"`
}

type DeletionDto struct {
	// The jira issue to use for committing the deletion.
	JiraIssue string `yaml:"-" json:"jiraIssue"`
}

type ErrorDto struct {
	Details   *string    `yaml:"details,omitempty" json:"details,omitempty"`
	Message   *string    `yaml:"message,omitempty" json:"message,omitempty"`
	Timestamp *time.Time `yaml:"timestamp,omitempty" json:"timestamp,omitempty"`
}

type HealthComponent struct {
	Description *string `yaml:"description,omitempty" json:"description,omitempty"`
	Status      *string `yaml:"status,omitempty" json:"status,omitempty"`
}

type Notification struct {
	// name of the service that was updated
	Name    string               `yaml:"name" json:"name"`
	Event   string               `yaml:"event" json:"event"`
	Type    string               `yaml:"type" json:"type"`
	Payload *NotificationPayload `yaml:"payload,omitempty" json:"payload,omitempty"`
}

type NotificationPayload struct {
	Owner      *OwnerDto      `yaml:"Owner,omitempty" json:"Owner,omitempty"`
	Service    *ServiceDto    `yaml:"Service,omitempty" json:"Service,omitempty"`
	Repository *RepositoryDto `yaml:"Repository,omitempty" json:"Repository,omitempty"`
}

type OwnerCreateDto struct {
	// The contact information of the owner
	Contact string `yaml:"contact" json:"contact"`
	// The teams channel url information of the owner
	TeamsChannelURL *string `yaml:"teamsChannelURL,omitempty" json:"teamsChannelURL,omitempty"`
	// The product owner of this owner space
	ProductOwner *string `yaml:"productOwner,omitempty" json:"productOwner,omitempty"`
	// A list of users that are allowed to promote services in this owner space
	Promoters []string `yaml:"promoters,omitempty" json:"promoters,omitempty"`
	// Map of string (group name e.g. some-owner) of strings (list of usernames), one username for each group is required.
	Groups *map[string][]string `yaml:"groups,omitempty" json:"groups,omitempty"`
	// The default jira project that is used by this owner space
	DefaultJiraProject *string `yaml:"defaultJiraProject,omitempty" json:"defaultJiraProject,omitempty"`
	// The jira issue to use for committing a change, or the last jira issue used.
	JiraIssue string `yaml:"-" json:"jiraIssue"`
	// A display name of the owner, to be presented in user interfaces instead of the owner's name, when available
	DisplayName *string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
}

type OwnerDto struct {
	// The contact information of the owner
	Contact string `yaml:"contact" json:"contact"`
	// The teams channel url information of the owner
	TeamsChannelURL *string `yaml:"teamsChannelURL,omitempty" json:"teamsChannelURL,omitempty"`
	// The product owner of this owner space
	ProductOwner *string `yaml:"productOwner,omitempty" json:"productOwner,omitempty"`
	// Map of string (group name e.g. some-owner) of strings (list of usernames), one username for each group is required.
	Groups *map[string][]string `yaml:"groups,omitempty" json:"groups,omitempty"`
	// A list of users that are allowed to promote services in this owner space
	Promoters []string `yaml:"promoters,omitempty" json:"promoters,omitempty"`
	// The default jira project that is used by this owner space
	DefaultJiraProject *string `yaml:"defaultJiraProject,omitempty" json:"defaultJiraProject,omitempty"`
	// ISO-8601 UTC date time at which this information was originally committed. When sending an update, include the original timestamp you got so we can detect concurrent updates.
	TimeStamp string `yaml:"-" json:"timeStamp"`
	// The git commit hash this information was originally committed under. When sending an update, include the original commitHash you got so we can detect concurrent updates.
	CommitHash string `yaml:"-" json:"commitHash"`
	// The jira issue to use for committing a change, or the last jira issue used.
	JiraIssue string `yaml:"-" json:"jiraIssue"`
	// A display name of the owner, to be presented in user interfaces instead of the owner's name, when available
	DisplayName *string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
}

type OwnerListDto struct {
	Owners map[string]OwnerDto `yaml:"owners" json:"owners"`
	// ISO-8601 UTC date time at which the list of owners was obtained from service-metadata
	TimeStamp string `yaml:"-" json:"timeStamp"`
}

type OwnerPatchDto struct {
	// The contact information of the owner
	Contact *string `yaml:"contact,omitempty" json:"contact,omitempty"`
	// The teams channel url information of the owner
	TeamsChannelURL *string `yaml:"teamsChannelURL,omitempty" json:"teamsChannelURL,omitempty"`
	// The product owner of this owner space
	ProductOwner *string `yaml:"productOwner,omitempty" json:"productOwner,omitempty"`
	// Map of string (group name e.g. some-owner) of strings (list of usernames), one username for each group is required.
	Groups *map[string][]string `yaml:"groups,omitempty" json:"groups,omitempty"`
	// A list of users that are allowed to promote services in this owner space
	Promoters []string `yaml:"promoters,omitempty" json:"promoters,omitempty"`
	// The default jira project that is used by this owner space
	DefaultJiraProject *string `yaml:"defaultJiraProject,omitempty" json:"defaultJiraProject,omitempty"`
	// ISO-8601 UTC date time at which this information was originally committed. When sending an update, include the original timestamp you got so we can detect concurrent updates.
	TimeStamp string `yaml:"-" json:"timeStamp"`
	// The git commit hash this information was originally committed under. When sending an update, include the original commitHash you got so we can detect concurrent updates.
	CommitHash string `yaml:"-" json:"commitHash"`
	// The jira issue to use for committing a change, or the last jira issue used.
	JiraIssue string `yaml:"-" json:"jiraIssue"`
	// A display name of the owner, to be presented in user interfaces instead of the owner's name, when available
	DisplayName *string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
}

type Quicklink struct {
	Url         *string `yaml:"url,omitempty" json:"url,omitempty"`
	Title       *string `yaml:"title,omitempty" json:"title,omitempty"`
	Description *string `yaml:"description,omitempty" json:"description,omitempty"`
}

type RepositoryConfigurationAccessKeyDto struct {
	Key        string  `yaml:"key" json:"key"`
	Permission *string `yaml:"permission,omitempty" json:"permission,omitempty"`
}

type RepositoryConfigurationDto struct {
	// Ssh-Keys configured on the repository.
	AccessKeys []RepositoryConfigurationAccessKeyDto `yaml:"accessKeys,omitempty" json:"accessKeys,omitempty"`
	// Adds a corresponding commit message regex.
	CommitMessageType *string `yaml:"commitMessageType,omitempty" json:"commitMessageType,omitempty"`
	// Configures JQL matcher with query: issuetype in (Story, Bug) AND 'Risk Level' is not EMPTY
	RequireIssue *bool `yaml:"requireIssue,omitempty" json:"requireIssue,omitempty"`
	// Set the required successful builds counter.
	RequireSuccessfulBuilds *int32 `yaml:"requireSuccessfulBuilds,omitempty" json:"requireSuccessfulBuilds,omitempty"`
	// Configuration of conditional builds as map of structs (key name e.g. some-key) of target references.
	RequireConditions *map[string]ConditionReferenceDto   `yaml:"requireConditions,omitempty" json:"requireConditions,omitempty"`
	Webhooks          *RepositoryConfigurationWebhooksDto `yaml:"webhooks,omitempty" json:"webhooks,omitempty"`
	// Map of string (group name e.g. some-owner) of strings (list of approvers), one approval for each group is required.
	Approvers *map[string][]string `yaml:"approvers,omitempty" json:"approvers,omitempty"`
	// List of strings (list of watchers, either usernames or group identifier), which are added as reviewers but require no approval.
	Watchers         []string `yaml:"watchers,omitempty" json:"watchers,omitempty"`
	DefaultReviewers []string `yaml:"defaultReviewers,omitempty" json:"defaultReviewers,omitempty"`
	// List of users, who can sign a pull request.
	SignedApprovers []string `yaml:"signedApprovers,omitempty" json:"signedApprovers,omitempty"`
	// Moves the repository into the archive.
	Archived *bool `yaml:"archived,omitempty" json:"archived,omitempty"`
	// Repository will not be configured, also not archived.
	Unmanaged *bool `yaml:"unmanaged,omitempty" json:"unmanaged,omitempty"`
}

type RepositoryConfigurationWebhookDto struct {
	Name string `yaml:"name" json:"name"`
	Url  string `yaml:"url" json:"url"`
	// Events the webhook should be triggered with.
	Events        []string           `yaml:"events,omitempty" json:"events,omitempty"`
	Configuration *map[string]string `yaml:"configuration,omitempty" json:"configuration,omitempty"`
}

type RepositoryConfigurationWebhooksDto struct {
	// Default pipeline trigger webhook.
	PipelineTrigger *bool `yaml:"pipelineTrigger,omitempty" json:"pipelineTrigger,omitempty"`
	// List of predefined webhooks
	Predefined []string `yaml:"predefined,omitempty" json:"predefined,omitempty"`
	// Additional webhooks to be configured.
	Additional []RepositoryConfigurationWebhookDto `yaml:"additional,omitempty" json:"additional,omitempty"`
}

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
	// The jira issue to use for committing a change, or the last jira issue used.
	JiraIssue string `yaml:"-" json:"jiraIssue"`
}

type RepositoryDto struct {
	// The alias of the repository owner
	Owner    string `yaml:"-" json:"owner"`
	Url      string `yaml:"url" json:"url"`
	Mainline string `yaml:"mainline" json:"mainline"`
	// the generator used for the initial contents of this repository
	Generator *string `yaml:"generator,omitempty" json:"generator,omitempty"`
	// this repository contains unit tests (currently ignored except for helm charts)
	Unittest      *bool                       `yaml:"unittest,omitempty" json:"unittest,omitempty"`
	Configuration *RepositoryConfigurationDto `yaml:"configuration,omitempty" json:"configuration,omitempty"`
	// ISO-8601 UTC date time at which this information was originally committed. When sending an update, include the original timestamp you got so we can detect concurrent updates.
	TimeStamp string `yaml:"-" json:"timeStamp"`
	// The git commit hash this information was originally committed under. When sending an update, include the original commitHash you got so we can detect concurrent updates.
	CommitHash string `yaml:"-" json:"commitHash"`
	// The jira issue to use for committing a change, or the last jira issue used.
	JiraIssue string `yaml:"-" json:"jiraIssue"`
}

type RepositoryListDto struct {
	Repositories map[string]RepositoryDto `yaml:"repositories" json:"repositories"`
	// ISO-8601 UTC date time at which the list of repositories was obtained from service-metadata
	TimeStamp string `yaml:"-" json:"timeStamp"`
}

type RepositoryPatchDto struct {
	// The alias of the repository owner
	Owner    *string `yaml:"owner,omitempty" json:"owner,omitempty"`
	Url      *string `yaml:"url,omitempty" json:"url,omitempty"`
	Mainline *string `yaml:"mainline,omitempty" json:"mainline,omitempty"`
	// the generator used for the initial contents of this repository
	Generator *string `yaml:"generator,omitempty" json:"generator,omitempty"`
	// this repository contains unit tests (currently ignored except for helm charts)
	Unittest      *bool                       `yaml:"unittest,omitempty" json:"unittest,omitempty"`
	Configuration *RepositoryConfigurationDto `yaml:"configuration,omitempty" json:"configuration,omitempty"`
	// ISO-8601 UTC date time at which this information was originally committed. When sending an update, include the original timestamp you got so we can detect concurrent updates.
	TimeStamp string `yaml:"-" json:"timeStamp"`
	// The git commit hash this information was originally committed under. When sending an update, include the original commitHash you got so we can detect concurrent updates.
	CommitHash string `yaml:"-" json:"commitHash"`
	// The jira issue to use for committing a change, or the last jira issue used.
	JiraIssue string `yaml:"-" json:"jiraIssue"`
}

type ServiceCreateDto struct {
	// The alias of the service owner. Note, an update with changed owner will move the service and any associated repositories to the new owner, but of course this will not move e.g. Jenkins jobs. That's your job.
	Owner string `yaml:"-" json:"owner"`
	// A short description of the functionality of the service.
	Description *string `yaml:"description,omitempty" json:"description,omitempty"`
	// A list of quicklinks related to the service
	Quicklinks []Quicklink `yaml:"quicklinks" json:"quicklinks"`
	// The keys of repositories associated with the service. When sending an update, they must refer to repositories that belong to this service, or the update will fail
	Repositories []string `yaml:"repositories" json:"repositories"`
	// The default channel used to send any alerts of the service to. Can be an email address or a Teams webhook URL
	AlertTarget string `yaml:"alertTarget" json:"alertTarget"`
	// True for services that will be permanently deployed to the Development environment only.
	DevelopmentOnly *bool `yaml:"developmentOnly,omitempty" json:"developmentOnly,omitempty"`
	// The operation type of the service. 'WORKLOAD' follows the default deployment strategy of one instance per environment, 'PLATFORM' one instance per cluster or node and 'APPLICATION' is a standalone application that is not deployed via the common strategies.
	OperationType *string `yaml:"operationType,omitempty" json:"operationType,omitempty"`
	// The value defines if the service is available from the internet and the time period in which security holes must be processed.
	InternetExposed *bool `yaml:"internetExposed,omitempty" json:"internetExposed,omitempty"`
	// The jira issue to use for committing a change, or the last jira issue used.
	JiraIssue string `yaml:"-" json:"jiraIssue"`
}

type ServiceDto struct {
	// The alias of the service owner. Note, an update with changed owner will move the service and any associated repositories to the new owner, but of course this will not move e.g. Jenkins jobs. That's your job.
	Owner string `yaml:"-" json:"owner"`
	// A short description of the functionality of the service.
	Description *string `yaml:"description,omitempty" json:"description,omitempty"`
	// A list of quicklinks related to the service
	Quicklinks []Quicklink `yaml:"quicklinks" json:"quicklinks"`
	// The keys of repositories associated with the service. When sending an update, they must refer to repositories that belong to this service, or the update will fail
	Repositories []string `yaml:"repositories" json:"repositories"`
	// The default channel used to send any alerts of the service to. Can be an email address or a Teams webhook URL
	AlertTarget string `yaml:"alertTarget" json:"alertTarget"`
	// True for services that will be permanently deployed to the Development environment only.
	DevelopmentOnly *bool `yaml:"developmentOnly,omitempty" json:"developmentOnly,omitempty"`
	// The operation type of the service. 'WORKLOAD' follows the default deployment strategy of one instance per environment, 'PLATFORM' one instance per cluster or node and 'APPLICATION' is a standalone application that is not deployed via the common strategies.
	OperationType *string `yaml:"operationType,omitempty" json:"operationType,omitempty"`
	// The value defines if the service is available from the internet and the time period in which security holes must be processed.
	InternetExposed *bool `yaml:"internetExposed,omitempty" json:"internetExposed,omitempty"`
	// ISO-8601 UTC date time at which this information was originally committed. When sending an update, include the original timestamp you got so we can detect concurrent updates.
	TimeStamp string `yaml:"-" json:"timeStamp"`
	// The git commit hash this information was originally committed under. When sending an update, include the original commitHash you got so we can detect concurrent updates.
	CommitHash string `yaml:"-" json:"commitHash"`
	// The jira issue to use for committing a change, or the last jira issue used.
	JiraIssue string `yaml:"-" json:"jiraIssue"`
	// The current phase of the service's development. A service usually starts off as 'experimental', then becomes 'operational' (i. e. can be reliably used and/or consumed). Once 'deprecated', the service doesn’t guarantee reliable use/consumption any longer.
	Lifecycle *string `yaml:"lifecycle,omitempty" json:"lifecycle,omitempty"`
}

type ServiceListDto struct {
	Services map[string]ServiceDto `yaml:"services" json:"services"`
	// ISO-8601 UTC date time at which the list of services was obtained from service-metadata
	TimeStamp string `yaml:"-" json:"timeStamp"`
}

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
	// True for services that will be permanently deployed to the Development environment only.
	DevelopmentOnly *bool `yaml:"developmentOnly,omitempty" json:"developmentOnly,omitempty"`
	// The operation type of the service. 'WORKLOAD' follows the default deployment strategy of one instance per environment, 'PLATFORM' one instance per cluster or node and 'APPLICATION' is a standalone application that is not deployed via the common strategies.
	OperationType *string `yaml:"operationType,omitempty" json:"operationType,omitempty"`
	// The value defines if the service is available from the internet and the time period in which security holes must be processed.
	InternetExposed *bool `yaml:"internetExposed,omitempty" json:"internetExposed,omitempty"`
	// ISO-8601 UTC date time at which this information was originally committed. When sending an update, include the original timestamp you got so we can detect concurrent updates.
	TimeStamp string `yaml:"-" json:"timeStamp"`
	// The git commit hash this information was originally committed under. When sending an update, include the original commitHash you got so we can detect concurrent updates.
	CommitHash string `yaml:"-" json:"commitHash"`
	// The jira issue to use for committing a change, or the last jira issue used.
	JiraIssue string `yaml:"-" json:"jiraIssue"`
	// The current phase of the service's development. A service usually starts off as 'experimental', then becomes 'operational' (i. e. can be reliably used and/or consumed). Once 'deprecated', the service doesn’t guarantee reliable use/consumption any longer.
	Lifecycle *string `yaml:"lifecycle,omitempty" json:"lifecycle,omitempty"`
}

type ServicePromotersDto struct {
	Promoters []string `yaml:"promoters" json:"promoters"`
}
