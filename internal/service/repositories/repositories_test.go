package repositories

import (
	"context"
	openapi "github.com/Interhyp/metadata-service/api/v1"
	"github.com/Interhyp/metadata-service/docs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func p(v string) *string {
	return &v
}

func b(v bool) *bool {
	return &v
}

func i(v int32) *int32 {
	return &v
}

func createRepositoryDto() openapi.RepositoryDto {
	return openapi.RepositoryDto{
		Owner:     "owner",
		Url:       "url",
		Mainline:  "mainline",
		Generator: p("generator"),
		Unittest:  b(true),
		Configuration: &openapi.RepositoryConfigurationDto{
			AccessKeys: []openapi.RepositoryConfigurationAccessKeyDto{
				{
					Key:        "KEY",
					Permission: p("REPO_WRITE"),
				},
			},
			CommitMessageType:       p("SEMANTIC"),
			RequireIssue:            b(false),
			RequireSuccessfulBuilds: i(1),
			Webhooks: &openapi.RepositoryConfigurationWebhooksDto{
				PipelineTrigger: b(false),
				Additional: []openapi.RepositoryConfigurationWebhookDto{
					{
						Name:          "webhookname",
						Url:           "webhookurl",
						Events:        []string{"event"},
						Configuration: &map[string]string{"key": "value"},
					},
				},
			},
			Approvers:        &map[string][]string{"group": {"approver1"}},
			DefaultReviewers: []string{"defaultreviewer1"},
			SignedApprovers:  []string{"signedapprover1"},
		},
		TimeStamp:  "ts",
		CommitHash: "hash",
	}
}

func createRepositoryDtoWithoutConfig() openapi.RepositoryDto {
	dto := createRepositoryDto()
	dto.Configuration = nil
	return dto
}

func assertPatchRepository(t *testing.T, current openapi.RepositoryDto, patch openapi.RepositoryPatchDto, expected openapi.RepositoryDto) {
	t.Helper()
	actual := patchRepository(current, patch)
	assert.Equal(t, expected, actual)
}

func TestPatchRepository_EmptyPatch(t *testing.T) {
	docs.Description("patching of repositories works with an empty patch")
	expected := createRepositoryDto()
	expected.TimeStamp = "newts"
	expected.CommitHash = "newhash"
	assertPatchRepository(t, createRepositoryDto(), openapi.RepositoryPatchDto{
		TimeStamp:  "newts",
		CommitHash: "newhash",
	}, expected)
}

func TestPatchRepository_WithConfig_And_EmptyOriginal(t *testing.T) {
	docs.Description("patching of repositories works with a missing original configuration")
	expected := createRepositoryDto()
	assertPatchRepository(t,
		createRepositoryDtoWithoutConfig(),
		openapi.RepositoryPatchDto{
			Configuration: expected.Configuration,
			TimeStamp:     expected.TimeStamp,
			CommitHash:    expected.CommitHash,
		},
		expected,
	)
}

func TestPatchRepository_ReplaceAll(t *testing.T) {
	docs.Description("patching of repositories works with an all fields patch")
	assertPatchRepository(t, createRepositoryDto(), openapi.RepositoryPatchDto{
		Owner:     p("newowner"),
		Url:       p("newurl"),
		Mainline:  p("newmainline"),
		Generator: p("newgenerator"),
		Unittest:  b(false),
		Configuration: &openapi.RepositoryConfigurationDto{
			AccessKeys: []openapi.RepositoryConfigurationAccessKeyDto{
				{
					Key:        "DEPLOYMENT",
					Permission: p("REPO_READ"),
				},
			},
			CommitMessageType:       p("DEFAULT"),
			RequireIssue:            b(true),
			RequireSuccessfulBuilds: i(2),
			Webhooks: &openapi.RepositoryConfigurationWebhooksDto{
				PipelineTrigger: b(false),
				Additional: []openapi.RepositoryConfigurationWebhookDto{
					{
						Name:          "newwebhookname",
						Url:           "newwebhookurl",
						Events:        []string{"event"},
						Configuration: &map[string]string{"newkey": "newvalue"},
					},
				},
			},
			Approvers:        &map[string][]string{"group": {"newapprover1"}},
			DefaultReviewers: []string{"newdefaultreviewer1"},
			SignedApprovers:  []string{"newsignedapprover1"},
		},
		TimeStamp:  "newts",
		CommitHash: "newhash",
	}, openapi.RepositoryDto{
		Owner:     "newowner",
		Url:       "newurl",
		Mainline:  "newmainline",
		Generator: p("newgenerator"),
		Unittest:  b(false),
		Configuration: &openapi.RepositoryConfigurationDto{
			AccessKeys: []openapi.RepositoryConfigurationAccessKeyDto{
				{
					Key:        "DEPLOYMENT",
					Permission: p("REPO_READ"),
				},
			},
			CommitMessageType:       p("DEFAULT"),
			RequireIssue:            b(true),
			RequireSuccessfulBuilds: i(2),
			Webhooks: &openapi.RepositoryConfigurationWebhooksDto{
				PipelineTrigger: b(false),
				Additional: []openapi.RepositoryConfigurationWebhookDto{
					{
						Name:          "newwebhookname",
						Url:           "newwebhookurl",
						Events:        []string{"event"},
						Configuration: &map[string]string{"newkey": "newvalue"},
					},
				},
			},
			Approvers:        &map[string][]string{"group": {"newapprover1"}},
			DefaultReviewers: []string{"newdefaultreviewer1"},
			SignedApprovers:  []string{"newsignedapprover1"},
		},
		TimeStamp:  "newts",
		CommitHash: "newhash",
	})
}

func TestPatchRepository_ClearFields(t *testing.T) {
	docs.Description("patching of repositories works with a patch that clears fields (result would not validate)")
	assertPatchRepository(t, createRepositoryDto(), openapi.RepositoryPatchDto{
		Owner:     p(""),
		Url:       p(""),
		Mainline:  p(""),
		Generator: p(""),
		Configuration: &openapi.RepositoryConfigurationDto{
			AccessKeys:        []openapi.RepositoryConfigurationAccessKeyDto{},
			CommitMessageType: p(""),
			Webhooks: &openapi.RepositoryConfigurationWebhooksDto{
				Additional: []openapi.RepositoryConfigurationWebhookDto{},
			},
			Approvers:        &map[string][]string{},
			DefaultReviewers: []string{},
			SignedApprovers:  []string{},
		},
		TimeStamp:  "",
		CommitHash: "",
	}, openapi.RepositoryDto{
		Owner:     "",
		Url:       "",
		Mainline:  "",
		Generator: nil,
		Unittest:  b(true),
		Configuration: &openapi.RepositoryConfigurationDto{
			AccessKeys:              nil,
			CommitMessageType:       nil,
			RequireIssue:            b(false),
			RequireSuccessfulBuilds: i(1),
			Webhooks: &openapi.RepositoryConfigurationWebhooksDto{
				PipelineTrigger: b(false),
				Additional:      nil,
			},
			Approvers:        nil,
			DefaultReviewers: nil,
			SignedApprovers:  nil,
		},
		TimeStamp:  "",
		CommitHash: "",
	})
}

func tstValid() openapi.RepositoryDto {
	return openapi.RepositoryDto{
		Owner:      "some-owner",
		Url:        "ssh://git@git.git:7999/GIT/git.git",
		Mainline:   "develop",
		TimeStamp:  "timestamp",
		CommitHash: "commithash",
		JiraIssue:  "jiraissue",
	}
}

func tstCreateValid() openapi.RepositoryCreateDto {
	return openapi.RepositoryCreateDto{
		Owner:     "some-owner",
		Url:       "ssh://git@git.git:7999/GIT/git.git",
		Mainline:  "develop",
		JiraIssue: "jiraissue",
	}
}

func tstValidationTestcaseAllOps(t *testing.T, expectedMessage string, data openapi.RepositoryDto, create openapi.RepositoryCreateDto, patch openapi.RepositoryPatchDto) {
	err := validateRepositoryCreateDto(context.TODO(), "any", create)
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, err.Error())

	err = validateExistingRepositoryDto(context.TODO(), "any", data)
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, err.Error())

	patch.TimeStamp = "newts"
	patch.CommitHash = "newhash"
	patch.JiraIssue = "newjiraissue"

	err = validateRepositoryPatchDto(context.TODO(), "any", patch, data)
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, err.Error())
}

func TestValidate_Url(t *testing.T) {
	docs.Description("invalid urls are correctly rejected on all operations")

	data := tstValid()
	data.Url = "https://no.this.is.not.correct.git"

	create := tstCreateValid()
	create.Url = "https://no.this.is.not.correct.git"

	patch := openapi.RepositoryPatchDto{
		Url: p("https://no.this.is.not.correct.git"),
	}

	expectedMessage := "validation error: field url must contain ssh git url"

	tstValidationTestcaseAllOps(t, expectedMessage, data, create, patch)
}

func TestValidate_Mainline(t *testing.T) {
	docs.Description("invalid urls are correctly rejected on all operations")

	data := tstValid()
	data.Mainline = "feature/hello"

	create := tstCreateValid()
	create.Mainline = "feature/hello"

	patch := openapi.RepositoryPatchDto{
		Mainline: p("feature/hello"),
	}

	expectedMessage := "validation error: mainline must be one of master, main, develop"

	tstValidationTestcaseAllOps(t, expectedMessage, data, create, patch)

	data.Mainline = ""
	create.Mainline = ""
	patch.Mainline = p("")

	expectedMessage = "validation error: field mainline is mandatory"

	tstValidationTestcaseAllOps(t, expectedMessage, data, create, patch)
}
