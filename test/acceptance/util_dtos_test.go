package acceptance

import (
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/repository/notifier"
)

func p(v string) *string {
	return &v
}

func pb(v bool) *bool {
	return &v
}

func pi(v int32) *int32 {
	return &v
}

// owner

func tstOwner() openapi.OwnerDto {
	return openapi.OwnerDto{
		Contact:            "somebody@some-organisation.com",
		TeamsChannelURL:    p("https://teams.microsoft.com/l/channel/somechannel"),
		ProductOwner:       p("kschlangenheld"),
		DefaultJiraProject: p("JIRA"),
		TimeStamp:          "2022-11-06T18:14:10Z",
		CommitHash:         "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:          "ISSUE-2345",
	}
}

func tstOwnerPatch() openapi.OwnerPatchDto {
	return openapi.OwnerPatchDto{
		Contact:         p("changed@some-organisation.com"),
		TeamsChannelURL: p("https://teams.microsoft.com/l/channel/somechannel"),
		TimeStamp:       "2022-11-06T18:14:10Z",
		CommitHash:      "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:       "ISSUE-2345",
	}
}

func tstOwnerUnchanged() openapi.OwnerDto {
	return openapi.OwnerDto{
		Contact:            "somebody@some-organisation.com",
		TeamsChannelURL:    p("https://teams.microsoft.com/l/channel/somechannel"),
		ProductOwner:       p("kschlangenheldt"),
		DefaultJiraProject: p("ISSUE"),
		TimeStamp:          "2022-11-06T18:14:10Z",
		CommitHash:         "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:          "ISSUE-2345",
	}
}

func tstOwnerUnchangedPatch() openapi.OwnerPatchDto {
	return openapi.OwnerPatchDto{
		Contact:         p("somebody@some-organisation.com"),
		TeamsChannelURL: p("https://teams.microsoft.com/l/channel/somechannel"),
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
defaultJiraProject: ISSUE
`
}

func tstOwnerPatchExpectedYaml() string {
	return `contact: changed@some-organisation.com
teamsChannelURL: https://teams.microsoft.com/l/channel/somechannel
productOwner: kschlangenheldt
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
			Url:   p("/swagger-ui/index.html"),
			Title: p("Swagger UI"),
		}},
		Repositories: []string{
			name + ".helm-deployment",
			name + ".implementation",
		},
		AlertTarget:     "squad_nothing@some-organisation.com",
		DevelopmentOnly: pb(false),
		OperationType:   nil,
		TimeStamp:       "2022-11-06T18:14:10Z",
		CommitHash:      "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:       "ISSUE-2345",
		Lifecycle:       p("experimental"),
		InternetExposed: pb(true),
	}
}

func tstServiceUnchanged(name string) openapi.ServiceDto {
	return openapi.ServiceDto{
		Owner: "some-owner",
		Quicklinks: []openapi.Quicklink{{
			Url:   p("/swagger-ui/index.html"),
			Title: p("Swagger UI"),
		}},
		Repositories: []string{
			name + ".helm-deployment",
			name + ".implementation",
		},
		AlertTarget:     "https://webhook.com/9asdflk29d4m39g",
		DevelopmentOnly: pb(false),
		OperationType:   nil,
		TimeStamp:       "2022-11-06T18:14:10Z",
		CommitHash:      "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:       "ISSUE-2345",
	}
}

func tstServicePatch() openapi.ServicePatchDto {
	return openapi.ServicePatchDto{
		AlertTarget:     p("squad_nothing@some-organisation.com"),
		TimeStamp:       "2022-11-06T18:14:10Z",
		CommitHash:      "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:       "ISSUE-2345",
		Lifecycle:       p("experimental"),
		InternetExposed: pb(true),
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
developmentOnly: false
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
developmentOnly: false
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
	fc := map[string][]string{
		"cached-template": {"cached-templates/tpl1.yaml", "more/cached/templates/tpl2.yaml"},
	}
	return openapi.RepositoryDto{
		Owner:    "some-owner",
		Url:      "ssh://git@bitbucket.some-organisation.com:7999/helm/karma-wrapper.git",
		Mainline: "master",
		Unittest: pb(false),
		Configuration: &openapi.RepositoryConfigurationDto{
			AccessKeys: []openapi.RepositoryConfigurationAccessKeyDto{
				{
					Key:        p("KEY"),
					Permission: p("REPO_WRITE"),
				},
			},
			CommitMessageType:       p("SEMANTIC"),
			RequireIssue:            pb(false),
			RequireSuccessfulBuilds: pi(1),
			RequireConditions:       &map[string]openapi.ConditionReferenceDto{"snyk-key": {RefMatcher: "master"}},
			Webhooks: &openapi.RepositoryConfigurationWebhooksDto{
				Additional: []openapi.RepositoryConfigurationWebhookDto{
					{
						Name:   "webhookname",
						Url:    "webhookurl",
						Events: []string{"event"},
					},
				},
			},
			Approvers: &map[string][]string{"testing": {"some-user"}},
		},
		Filecategory: &fc,
		TimeStamp:    "2022-11-06T18:14:10Z",
		CommitHash:   "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:    "ISSUE-2345",
	}
}

func tstRepositoryUnchanged() openapi.RepositoryDto {
	return openapi.RepositoryDto{
		Owner:      "some-owner",
		Url:        "ssh://git@bitbucket.some-organisation.com:7999/helm/karma-wrapper.git",
		Mainline:   "master",
		Unittest:   pb(false),
		TimeStamp:  "2022-11-06T18:14:10Z",
		CommitHash: "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:  "ISSUE-2345",
	}
}

func tstRepositoryPatch() openapi.RepositoryPatchDto {
	return openapi.RepositoryPatchDto{
		Mainline:   p("main"),
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
    requireIssue: false
    requireSuccessfulBuilds: 1
    requireConditions:
        snyk-key:
            refMatcher: master
    webhooks:
        additional:
            - name: webhookname
              url: webhookurl
              events:
                - event
    approvers:
        testing:
            - some-user
filecategory:
    cached-template:
        - cached-templates/tpl1.yaml
        - more/cached/templates/tpl2.yaml
`
}

func tstRepositoryExpectedYamlKarmaWrapper() string {
	return `url: ssh://git@bitbucket.some-organisation.com:7999/helm/karma-wrapper.git
mainline: main
unittest: false
`
}

func tstRepositoryUnchangedExpectedYaml() string {
	return `url: ssh://git@bitbucket.some-organisation.com:7999/helm/karma-wrapper.git
mainline: master
unittest: false
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
	return notifier.AsPayload(repo)
}
