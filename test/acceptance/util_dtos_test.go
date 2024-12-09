package acceptance

import (
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/repository/notifier"
	"github.com/Interhyp/metadata-service/internal/util"
)

func ptr[T interface{}](v T) *T {
	return &v
}

// owner

func tstOwner() openapi.OwnerDto {
	return openapi.OwnerDto{
		Contact:            "somebody@some-organisation.com",
		TeamsChannelURL:    ptr("https://teams.microsoft.com/l/channel/somechannel"),
		ProductOwner:       ptr("kschlangenheld"),
		DefaultJiraProject: ptr("JIRA"),
		TimeStamp:          "2022-11-06T18:14:10Z",
		CommitHash:         "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:          "ISSUE-2345",
	}
}

func tstOwnerPatch() openapi.OwnerPatchDto {
	return openapi.OwnerPatchDto{
		Contact:         ptr("changed@some-organisation.com"),
		TeamsChannelURL: ptr("https://teams.microsoft.com/l/channel/somechannel"),
		TimeStamp:       "2022-11-06T18:14:10Z",
		CommitHash:      "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:       "ISSUE-2345",
	}
}

func tstOwnerUnchanged() openapi.OwnerDto {
	return openapi.OwnerDto{
		Contact:            "somebody@some-organisation.com",
		TeamsChannelURL:    ptr("https://teams.microsoft.com/l/channel/somechannel"),
		ProductOwner:       ptr("kschlangenheldt"),
		DefaultJiraProject: ptr("ISSUE"),
		TimeStamp:          "2022-11-06T18:14:10Z",
		CommitHash:         "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:          "ISSUE-2345",
		Groups: map[string][]string{
			"users": {
				"some-other-user",
				"a-very-special-user",
			},
		},
	}
}

func tstOwnerUnchangedPatch() openapi.OwnerPatchDto {
	return openapi.OwnerPatchDto{
		Contact:         ptr("somebody@some-organisation.com"),
		TeamsChannelURL: ptr("https://teams.microsoft.com/l/channel/somechannel"),
		TimeStamp:       "2022-11-06T18:14:10Z",
		CommitHash:      "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:       "ISSUE-2345",
	}
}

func tstNewOwnerPayload() openapi.NotificationPayload {
	owner := tstOwner()
	owner.CommitHash = "6c8ac2c35791edf9979623c717a2430000000000"
	return notifier.AsPayload(owner)
}

func tstUpdatedOwnerPayload() openapi.NotificationPayload {
	owner := tstOwnerUnchanged()
	owner.CommitHash = "6c8ac2c35791edf9979623c717a2430000000000"
	owner.Contact = "changed@some-organisation.com"
	return notifier.AsPayload(owner)
}

func tstOwnerExpectedYaml() string {
	return `contact: somebody@some-organisation.com
teamsChannelURL: https://teams.microsoft.com/l/channel/somechannel
productOwner: kschlangenheld
defaultJiraProject: JIRA
`
}

func tstOwnerUnchangedExpectedYaml() string {
	return `contact: somebody@some-organisation.com
teamsChannelURL: https://teams.microsoft.com/l/channel/somechannel
productOwner: kschlangenheldt
groups:
    users:
        - some-other-user
        - a-very-special-user
defaultJiraProject: ISSUE
`
}

func tstOwnerPatchExpectedYaml() string {
	return `contact: changed@some-organisation.com
teamsChannelURL: https://teams.microsoft.com/l/channel/somechannel
productOwner: kschlangenheldt
groups:
    users:
        - some-other-user
        - a-very-special-user
defaultJiraProject: ISSUE
`
}

func tstOwnerExpectedKafka(alias string) string {
	return `{"affected":{"ownerAliases":["` + alias +
		`"],"serviceNames":[],"repositoryKeys":[]},"timeStamp":"2022-11-06T18:14:10Z",` +
		`"commitHash":"6c8ac2c35791edf9979623c717a2430000000000"}`
}

// service

func tstService(name string) openapi.ServiceDto {
	return openapi.ServiceDto{
		Owner: "some-owner",
		Quicklinks: []openapi.Quicklink{{
			Url:   ptr("/swagger-ui/index.html"),
			Title: ptr("Swagger UI"),
		}},
		Repositories: []string{
			name + ".helm-deployment",
			name + ".implementation",
		},
		AlertTarget:     "squad_nothing@some-organisation.com",
		OperationType:   nil,
		TimeStamp:       "2022-11-06T18:14:10Z",
		CommitHash:      "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:       "ISSUE-2345",
		Lifecycle:       ptr("experimental"),
		InternetExposed: ptr(true),
	}
}

func tstServiceUnchanged(name string) openapi.ServiceDto {
	return openapi.ServiceDto{
		Owner: "some-owner",
		Quicklinks: []openapi.Quicklink{{
			Url:   ptr("/swagger-ui/index.html"),
			Title: ptr("Swagger UI"),
		}},
		Repositories: []string{
			name + ".helm-deployment",
			name + ".implementation",
		},
		AlertTarget:   "https://webhook.com/9asdflk29d4m39g",
		OperationType: nil,
		TimeStamp:     "2022-11-06T18:14:10Z",
		CommitHash:    "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:     "ISSUE-2345",
	}
}

func tstServicePatch() openapi.ServicePatchDto {
	return openapi.ServicePatchDto{
		AlertTarget:     ptr("squad_nothing@some-organisation.com"),
		TimeStamp:       "2022-11-06T18:14:10Z",
		CommitHash:      "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:       "ISSUE-2345",
		Lifecycle:       ptr("experimental"),
		InternetExposed: ptr(true),
	}
}

func tstServiceUnchangedPatch() openapi.ServicePatchDto {
	return openapi.ServicePatchDto{
		TimeStamp:  "2022-11-06T18:14:10Z",
		CommitHash: "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:  "ISSUE-2345",
	}
}

func tstServiceExpectedYaml(name string) string {
	return `quicklinks:
    - url: /swagger-ui/index.html
      title: Swagger UI
repositories:
    - ` + name + `/helm-deployment
    - ` + name + `/implementation
alertTarget: squad_nothing@some-organisation.com
internetExposed: true
lifecycle: experimental
`
}

func tstServiceUnchangedExpectedYaml(name string) string {
	return `quicklinks:
    - url: /swagger-ui/index.html
      title: Swagger UI
repositories:
    - ` + name + `/helm-deployment
    - ` + name + `/implementation
alertTarget: https://webhook.com/9asdflk29d4m39g
`
}

func tstServiceExpectedKafka(name string) string {
	return `{"affected":{"ownerAliases":[],"serviceNames":["` +
		name + `"],"repositoryKeys":[]},"timeStamp":"2022-11-06T18:14:10Z",` +
		`"commitHash":"6c8ac2c35791edf9979623c717a2430000000000"}`
}

func tstServiceMovedExpectedKafka(name string) string {
	return `{"affected":{"ownerAliases":[],"serviceNames":["` +
		name + `"],"repositoryKeys":["` + name + `.helm-deployment","` +
		name + `.implementation"]},"timeStamp":"2022-11-06T18:14:10Z",` +
		`"commitHash":"6c8ac2c35791edf9979623c717a2430000000000"}`
}

func tstNewServicePayload(name string) openapi.NotificationPayload {
	service := tstService(name)
	service.CommitHash = "6c8ac2c35791edf9979623c717a2430000000000"
	return notifier.AsPayload(service)
}

func tstUpdatedServicePayload(name string) openapi.NotificationPayload {
	service := tstService(name)
	service.CommitHash = "6c8ac2c35791edf9979623c717a2430000000000"
	return notifier.AsPayload(service)
}

// repository

func tstRepository() openapi.RepositoryDto {
	return openapi.RepositoryDto{
		Owner:    "some-owner",
		Url:      "ssh://git@bitbucket.some-organisation.com:7999/helm/karma-wrapper.git",
		Mainline: "master",
		Unittest: ptr(false),
		Configuration: &openapi.RepositoryConfigurationDto{
			AccessKeys: []openapi.RepositoryConfigurationAccessKeyDto{
				{
					Key:        ptr("KEY"),
					Permission: ptr("REPO_WRITE"),
				},
			},
			CommitMessageType:       ptr("SEMANTIC"),
			RequireIssue:            ptr(false),
			RequireSuccessfulBuilds: ptr(int32(1)),
			RequireConditions:       map[string]openapi.ConditionReferenceDto{"snyk-key": {RefMatcher: "master"}},
			Webhooks: &openapi.RepositoryConfigurationWebhooksDto{
				Additional: []openapi.RepositoryConfigurationWebhookDto{
					{
						Name:   "webhookname",
						Url:    "webhookurl",
						Events: []string{"event"},
					},
				},
			},
			Approvers: map[string][]string{"testing": {"some-user"}},
		},
		TimeStamp:  "2022-11-06T18:14:10Z",
		CommitHash: "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:  "ISSUE-2345",
	}
}

func tstRepositoryUnchanged() openapi.RepositoryDto {
	return openapi.RepositoryDto{
		Owner:    "some-owner",
		Url:      "ssh://git@bitbucket.some-organisation.com:7999/helm/karma-wrapper.git",
		Mainline: "master",
		Unittest: ptr(false),
		Configuration: &openapi.RepositoryConfigurationDto{
			BranchNameRegex: ptr("testing_.*"),
		},
		TimeStamp:  "2022-11-06T18:14:10Z",
		CommitHash: "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:  "ISSUE-2345",
	}
}

func tstRepositoryPatch() openapi.RepositoryPatchDto {
	return openapi.RepositoryPatchDto{
		Mainline: ptr("main"),
		Configuration: &openapi.RepositoryConfigurationPatchDto{
			BranchNameRegex:   ptr("testing_.*"),
			RequireIssue:      ptr(true),
			RequireConditions: make(map[string]openapi.ConditionReferenceDto),
			RefProtections: &openapi.RefProtections{
				Branches: &openapi.RefProtectionsBranches{
					RequirePR: []openapi.ProtectedRef{
						{
							Pattern: ".*",
						},
					},
				},
			},
		},
		TimeStamp:  "2022-11-06T18:14:10Z",
		CommitHash: "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:  "ISSUE-2345",
	}
}

func tstRepositoryPatchWithIgnoredConfigurationFields() interface{} {
	return struct {
		Mainline      *string                             `yaml:"mainline,omitempty" json:"mainline,omitempty"`
		Configuration *openapi.RepositoryConfigurationDto `yaml:"configuration,omitempty" json:"configuration,omitempty"`
		TimeStamp     string                              `yaml:"-" json:"timeStamp"`
		CommitHash    string                              `yaml:"-" json:"commitHash"`
		JiraIssue     string                              `yaml:"-" json:"jiraIssue"`
		Labels        map[string]string                   `yaml:"labels,omitempty" json:"labels,omitempty"`
	}{
		Mainline: ptr("main"),
		Configuration: &openapi.RepositoryConfigurationDto{
			RequireIssue: ptr(true),
			RefProtections: &openapi.RefProtections{
				Branches: &openapi.RefProtectionsBranches{
					RequirePR: []openapi.ProtectedRef{
						{
							Pattern: ".*",
						},
					},
				},
			},
		},
		TimeStamp:  "2022-11-06T18:14:10Z",
		CommitHash: "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:  "ISSUE-2345",
	}
}

func tstRepositoryUnchangedPatch() openapi.RepositoryPatchDto {
	return openapi.RepositoryPatchDto{
		TimeStamp:  "2022-11-06T18:14:10Z",
		CommitHash: "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:  "ISSUE-2345",
	}
}

func tstRepositoryExpectedYaml() string {
	return `url: ssh://git@bitbucket.some-organisation.com:7999/helm/karma-wrapper.git
mainline: master
unittest: false
configuration:
    accessKeys:
        - key: KEY
          permission: REPO_WRITE
    commitMessageType: SEMANTIC
    requireSuccessfulBuilds: 1
    webhooks:
        additional:
            - name: webhookname
              url: webhookurl
              events:
                - event
    approvers:
        testing:
            - some-user
    requireIssue: false
    requireConditions:
        snyk-key:
            refMatcher: master
`
}

func tstRepositoryExpectedYamlKarmaWrapper() string {
	return `url: ssh://git@bitbucket.some-organisation.com:7999/helm/karma-wrapper.git
mainline: main
unittest: false
configuration:
    branchNameRegex: testing_.*
    refProtections:
        branches:
            requirePR:
                - pattern: .*
    requireIssue: true
`
}

func tstRepositoryUnchangedExpectedYaml() string {
	return `url: ssh://git@bitbucket.some-organisation.com:7999/helm/karma-wrapper.git
mainline: master
unittest: false
configuration:
    branchNameRegex: testing_.*
`
}

func tstRepositoryExpectedKafka(key string) string {
	return `{"affected":{"ownerAliases":[],"serviceNames":[],"repositoryKeys":["` +
		key + `"]},"timeStamp":"2022-11-06T18:14:10Z",` +
		`"commitHash":"6c8ac2c35791edf9979623c717a2430000000000"}`
}

func tstDelete() openapi.DeletionDto {
	return openapi.DeletionDto{
		JiraIssue: "ISSUE-2345",
	}
}

func tstNewRepositoryPayload() openapi.NotificationPayload {
	repo := tstRepository()
	repo.CommitHash = "6c8ac2c35791edf9979623c717a2430000000000"
	return notifier.AsPayload(repo)
}

func tstUpdatedRepositoryPayload() openapi.NotificationPayload {
	repo := tstRepositoryUnchanged()
	repo.Mainline = "main"
	repo.CommitHash = "6c8ac2c35791edf9979623c717a2430000000000"
	repo.Configuration = &openapi.RepositoryConfigurationDto{
		BranchNameRegex: ptr("testing_.*"),
		RequireIssue:    util.Ptr(true),
		RefProtections: &openapi.RefProtections{
			Branches: &openapi.RefProtectionsBranches{
				RequirePR: []openapi.ProtectedRef{
					{Pattern: ".*"},
				},
			},
		},
	}
	return notifier.AsPayload(repo)
}
