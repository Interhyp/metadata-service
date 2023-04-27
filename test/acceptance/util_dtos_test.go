package acceptance

import openapi "github.com/Interhyp/metadata-service/api/v1"

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
		ProductOwner:       p("kschlangenheld"),
		Promoters:          []string{"someone", "else", "entirely"},
		DefaultJiraProject: p("JIRA"),
		TimeStamp:          "2022-11-06T18:14:10Z",
		CommitHash:         "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:          "ISSUE-2345",
	}
}

func tstOwnerPatch() openapi.OwnerPatchDto {
	return openapi.OwnerPatchDto{
		Contact:    p("somebody@some-organisation.com"),
		Promoters:  []string{"someone", "else", "entirely"},
		TimeStamp:  "2022-11-06T18:14:10Z",
		CommitHash: "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:  "ISSUE-2345",
	}
}

func tstOwnerUnchanged() openapi.OwnerDto {
	return openapi.OwnerDto{
		Contact:            "somebody@some-organisation.com",
		ProductOwner:       p("kschlangenheldt"),
		Promoters:          []string{"someone", "else"},
		DefaultJiraProject: p("ISSUE"),
		TimeStamp:          "2022-11-06T18:14:10Z",
		CommitHash:         "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:          "ISSUE-2345",
	}
}

func tstOwnerUnchangedPatch() openapi.OwnerPatchDto {
	return openapi.OwnerPatchDto{
		Contact:    p("somebody@some-organisation.com"),
		TimeStamp:  "2022-11-06T18:14:10Z",
		CommitHash: "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:  "ISSUE-2345",
	}
}

func tstOwnerExpectedYaml() string {
	return `contact: somebody@some-organisation.com
productOwner: kschlangenheld
promoters:
    - someone
    - else
    - entirely
defaultJiraProject: JIRA
`
}

func tstOwnerUnchangedExpectedYaml() string {
	return `contact: somebody@some-organisation.com
productOwner: kschlangenheldt
promoters:
    - someone
    - else
defaultJiraProject: ISSUE
`
}

func tstOwnerPatchExpectedYaml() string {
	return `contact: somebody@some-organisation.com
productOwner: kschlangenheldt
promoters:
    - someone
    - else
    - entirely
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
		RequiredScans:   []string{"SAST", "SCA"},
		TimeStamp:       "2022-11-06T18:14:10Z",
		CommitHash:      "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:       "ISSUE-2345",
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
		RequiredScans:   []string{"SAST", "SCA"},
		TimeStamp:       "2022-11-06T18:14:10Z",
		CommitHash:      "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:       "ISSUE-2345",
	}
}

func tstServicePatch() openapi.ServicePatchDto {
	return openapi.ServicePatchDto{
		AlertTarget: p("squad_nothing@some-organisation.com"),
		TimeStamp:   "2022-11-06T18:14:10Z",
		CommitHash:  "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:   "ISSUE-2345",
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
requiredScans:
    - SAST
    - SCA
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
requiredScans:
    - SAST
    - SCA
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

// repository

func tstRepository() openapi.RepositoryDto {
	return openapi.RepositoryDto{
		Owner:    "some-owner",
		Url:      "ssh://git@bitbucket.some-organisation.com:7999/helm/karma-wrapper.git",
		Mainline: "master",
		Unittest: pb(false),
		Configuration: &openapi.RepositoryConfigurationDto{
			AccessKeys: []openapi.RepositoryConfigurationAccessKeyDto{
				{
					Key:        "KEY",
					Permission: p("REPO_WRITE"),
				},
			},
			CommitMessageType:       p("SEMANTIC"),
			RequireIssue:            pb(false),
			RequireSuccessfulBuilds: pi(1),
			RequireConditions:       &map[string]openapi.ConditionReferenceDto{"snyk-key": {RefMatcher: "master"}},
			Webhooks: &openapi.RepositoryConfigurationWebhooksDto{
				PipelineTrigger: pb(false),
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
		TimeStamp:  "2022-11-06T18:14:10Z",
		CommitHash: "6c8ac2c35791edf9979623c717a243fc53400000",
		JiraIssue:  "ISSUE-2345",
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
        pipelineTrigger: false
        additional:
            - name: webhookname
              url: webhookurl
              events:
                - event
    approvers:
        testing:
            - some-user
`
}

func tstRepositoryExpectedYamlKarmaWrapper() string {
	return `url: ssh://git@bitbucket.some-organisation.com:7999/helm/karma-wrapper.git
mainline: master
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
