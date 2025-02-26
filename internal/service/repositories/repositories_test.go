package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/Interhyp/go-backend-service-common/api/apierrors"
	"github.com/Interhyp/go-backend-service-common/docs"
	"github.com/Interhyp/go-backend-service-common/repository/timestamp"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	"github.com/Interhyp/metadata-service/internal/service/owners"
	auloggingapi "github.com/StephanHCB/go-autumn-logging/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T {
	return &v
}

func i(v int32) *int32 {
	return &v
}

func createRepositoryDto() openapi.RepositoryDto {
	return openapi.RepositoryDto{
		Owner:         "owner",
		Url:           "url",
		Mainline:      "mainline",
		Generator:     ptr("generator"),
		Configuration: createRepositoryConfigurationDto(),
		Labels:        map[string]string{"label": "originalValue"},
		TimeStamp:     "ts",
		CommitHash:    "hash",
	}
}

func createRepositoryConfigurationDto() *openapi.RepositoryConfigurationDto {
	return &openapi.RepositoryConfigurationDto{
		AccessKeys: []openapi.RepositoryConfigurationAccessKeyDto{
			{
				Key:        ptr("KEY"),
				Permission: ptr("REPO_WRITE"),
			},
		},
		CommitMessageType:       ptr("SEMANTIC"),
		RequireIssue:            ptr(false),
		RequireSuccessfulBuilds: i(1),
		Webhooks: &openapi.RepositoryConfigurationWebhooksDto{
			Additional: []openapi.RepositoryConfigurationWebhookDto{
				{
					Name:          "webhookname",
					Url:           "webhookurl",
					Events:        []string{"event"},
					Configuration: map[string]string{"key": "value"},
				},
			},
		},
		Approvers: map[string][]string{"group": {"approver1"}},
		Archived:  ptr(false),
	}
}
func createRepositoryConfigurationPatchDtoFromConfigurationDto(input *openapi.RepositoryConfigurationDto) *openapi.RepositoryConfigurationPatchDto {
	return &openapi.RepositoryConfigurationPatchDto{
		AccessKeys:              input.AccessKeys,
		CommitMessageType:       input.CommitMessageType,
		ActionsAccess:           input.ActionsAccess,
		RequireSuccessfulBuilds: input.RequireSuccessfulBuilds,
		Webhooks:                input.Webhooks,
		Approvers:               input.Approvers,
		Archived:                input.Archived,
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
	assert.EqualExportedValues(t, expected, actual)
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
	beforePatch := createRepositoryDtoWithoutConfig()
	expected := createRepositoryDto()
	beforePatch.Configuration = &openapi.RepositoryConfigurationDto{
		//transfer not patchable fields to beforePatch
		RefProtections:    expected.Configuration.RefProtections,
		RequireIssue:      expected.Configuration.RequireIssue,
		RequireConditions: expected.Configuration.RequireConditions,
	}
	assertPatchRepository(t,
		beforePatch,
		openapi.RepositoryPatchDto{
			//Generator:     expected.Generator,
			Configuration: createRepositoryConfigurationPatchDtoFromConfigurationDto(expected.Configuration),
			TimeStamp:     expected.TimeStamp,
			CommitHash:    expected.CommitHash,
		},
		expected,
	)
}

func TestPatchRepository_ReplaceAll(t *testing.T) {
	docs.Description("patching of repositories works with an all fields patch")
	assertPatchRepository(t, createRepositoryDto(), openapi.RepositoryPatchDto{
		Owner:     ptr("newowner"),
		Url:       ptr("newurl"),
		Mainline:  ptr("newmainline"),
		Generator: ptr("newgenerator"),
		Configuration: &openapi.RepositoryConfigurationPatchDto{
			AccessKeys: []openapi.RepositoryConfigurationAccessKeyDto{
				{
					Key:        ptr("DEPLOYMENT"),
					Permission: ptr("REPO_READ"),
				},
			},
			BranchNameRegex:         ptr("(testing_[^_-]+_[^-]+$)"),
			CommitMessageRegex:      ptr("(([A-Z][A-Z_0-9]+-[0-9]+))(.|\\n)*"),
			CommitMessageType:       ptr("DEFAULT"),
			ActionsAccess:           ptr("ENTERPRISE"),
			RequireSuccessfulBuilds: i(2),
			Webhooks: &openapi.RepositoryConfigurationWebhooksDto{
				Additional: []openapi.RepositoryConfigurationWebhookDto{
					{
						Name:          "newwebhookname",
						Url:           "newwebhookurl",
						Events:        []string{"event"},
						Configuration: map[string]string{"newkey": "newvalue"},
					},
				},
			},
			Approvers: map[string][]string{"group": {"newapprover1"}},
			Archived:  ptr(true),
		},
		Labels:     map[string]string{"label": "patchedValue"},
		TimeStamp:  "newts",
		CommitHash: "newhash",
	}, openapi.RepositoryDto{
		Owner:     "newowner",
		Url:       "newurl",
		Mainline:  "newmainline",
		Generator: ptr("newgenerator"),
		Configuration: &openapi.RepositoryConfigurationDto{
			AccessKeys: []openapi.RepositoryConfigurationAccessKeyDto{
				{
					Key:        ptr("DEPLOYMENT"),
					Permission: ptr("REPO_READ"),
				},
			},
			BranchNameRegex:         ptr("(testing_[^_-]+_[^-]+$)"),
			CommitMessageRegex:      ptr("(([A-Z][A-Z_0-9]+-[0-9]+))(.|\\n)*"),
			CommitMessageType:       ptr("DEFAULT"),
			ActionsAccess:           ptr("ENTERPRISE"),
			RequireIssue:            ptr(false),
			RequireSuccessfulBuilds: i(2),
			Webhooks: &openapi.RepositoryConfigurationWebhooksDto{
				Additional: []openapi.RepositoryConfigurationWebhookDto{
					{
						Name:          "newwebhookname",
						Url:           "newwebhookurl",
						Events:        []string{"event"},
						Configuration: map[string]string{"newkey": "newvalue"},
					},
				},
			},
			Approvers: map[string][]string{"group": {"newapprover1"}},
			Archived:  ptr(true),
		},
		Labels:     map[string]string{"label": "patchedValue"},
		TimeStamp:  "newts",
		CommitHash: "newhash",
	})
}

func TestPatchRepository_ClearFields(t *testing.T) {
	docs.Description("patching of repositories works with a patch that clears fields (result would not validate)")
	assertPatchRepository(t, createRepositoryDto(), openapi.RepositoryPatchDto{
		Owner:     ptr(""),
		Url:       ptr(""),
		Mainline:  ptr(""),
		Generator: ptr(""),
		Configuration: &openapi.RepositoryConfigurationPatchDto{
			AccessKeys:        []openapi.RepositoryConfigurationAccessKeyDto{},
			CommitMessageType: ptr(""),
			ActionsAccess:     ptr(""),
			Webhooks: &openapi.RepositoryConfigurationWebhooksDto{
				Additional: []openapi.RepositoryConfigurationWebhookDto{},
			},
			Approvers: map[string][]string{},
		},
		Labels:     map[string]string{},
		TimeStamp:  "",
		CommitHash: "",
	}, openapi.RepositoryDto{
		Owner:     "",
		Url:       "",
		Mainline:  "",
		Generator: nil,
		Configuration: &openapi.RepositoryConfigurationDto{
			AccessKeys:              nil,
			CommitMessageType:       nil,
			ActionsAccess:           nil,
			RequireIssue:            ptr(false),
			RequireSuccessfulBuilds: i(1),
			Webhooks: &openapi.RepositoryConfigurationWebhooksDto{
				Additional: nil,
			},
			Approvers: nil,
			Archived:  ptr(false),
		},
		Labels:     nil,
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

type MockLogging struct {
}

func (m MockLogging) IsLogging() bool {
	//TODO implement me
	panic("implement me")
}

func (m MockLogging) Setup() {
	//TODO implement me
	panic("implement me")
}

func (m MockLogging) Logger() auloggingapi.LoggingImplementation {
	return MockLoggingImplementation{}
}

type MockLoggingImplementation struct {
}

func (m MockLoggingImplementation) Ctx(ctx context.Context) auloggingapi.ContextAwareLoggingImplementation {
	return MockContextAwareLoggingImplementation{}
}

func (m MockLoggingImplementation) NoCtx() auloggingapi.ContextAwareLoggingImplementation {
	//TODO implement me
	panic("implement me")
}

type MockContextAwareLoggingImplementation struct {
}

func (m MockContextAwareLoggingImplementation) Trace() auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

func (m MockContextAwareLoggingImplementation) Debug() auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

func (m MockContextAwareLoggingImplementation) Info() auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

func (m MockContextAwareLoggingImplementation) Warn() auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

func (m MockContextAwareLoggingImplementation) Error() auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

func (m MockContextAwareLoggingImplementation) Fatal() auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

func (m MockContextAwareLoggingImplementation) Panic() auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

type MockLeveledLoggingImplementation struct {
}

func (m MockLeveledLoggingImplementation) WithErr(err error) auloggingapi.LeveledLoggingImplementation {
	//TODO implement me
	panic("implement me")
}

func (m MockLeveledLoggingImplementation) With(key string, value string) auloggingapi.LeveledLoggingImplementation {
	//TODO implement me
	panic("implement me")
}

func (m MockLeveledLoggingImplementation) Print(v ...interface{}) {
	//TODO implement me
	panic("implement me")
}

func (m MockLeveledLoggingImplementation) Printf(format string, v ...interface{}) {
	// do nothing
}

func fakeNow() time.Time {
	return time.Date(2022, 11, 6, 18, 14, 10, 0, time.UTC)
}

func tstValidationTestcaseAllOps(t *testing.T, expectedMessage string, data openapi.RepositoryDto, create openapi.RepositoryCreateDto, patch openapi.RepositoryPatchDto) {
	mockLogging := MockLogging{}
	fakeNow := func() time.Time {
		return time.Date(2022, 11, 6, 18, 14, 10, 0, time.UTC)
	}
	timestampImpl := timestamp.TimestampImpl{
		Timestamp: fakeNow,
	}
	impl := &Impl{
		Configuration: nil,
		Logging:       &mockLogging,
		Cache:         nil,
		Updater:       nil,
		Timestamp:     &timestampImpl,
	}

	err := (*Impl).validateRepositoryCreateDto(impl, context.TODO(), "any", create)
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, *err.(apierrors.AnnotatedError).ApiError().Details)

	err = (*Impl).validateExistingRepositoryDto(impl, context.TODO(), "any", data)
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, *err.(apierrors.AnnotatedError).ApiError().Details)

	patch.TimeStamp = "newts"
	patch.CommitHash = "newhash"
	patch.JiraIssue = "newjiraissue"

	err = (*Impl).validateRepositoryPatchDto(impl, context.TODO(), "any", patch, data)
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, *err.(apierrors.AnnotatedError).ApiError().Details)
}

func TestValidate_Url(t *testing.T) {
	docs.Description("invalid urls are correctly rejected on all operations")

	data := tstValid()
	data.Url = "https://no.this.is.not.correct.git"

	create := tstCreateValid()
	create.Url = "https://no.this.is.not.correct.git"

	patch := openapi.RepositoryPatchDto{
		Url: ptr("https://no.this.is.not.correct.git"),
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
		Mainline: ptr("feature/hello"),
	}

	expectedMessage := "validation error: mainline must be one of master, main, develop"

	tstValidationTestcaseAllOps(t, expectedMessage, data, create, patch)

	data.Mainline = ""
	create.Mainline = ""
	patch.Mainline = ptr("")

	expectedMessage = "validation error: field mainline is mandatory"

	tstValidationTestcaseAllOps(t, expectedMessage, data, create, patch)
}

func TestRebuildApprovers_DuplicatesAndMultipleGroups(t *testing.T) {
	instance := createInstance()

	testApprovers := make(map[string][]string, 0)
	testApprovers["one"] = []string{"x", "y", "z", "z"}
	testApprovers["two"] = []string{"z", "o", "v", "v"}
	configDto := createRepositoryConfigDto(testApprovers)

	instance.expandApprovers(context.TODO(), configDto.Approvers)

	require.Equal(t, 2, len(configDto.Approvers))
	require.Exactly(t, configDto.Approvers["one"], []string{"x", "y", "z"})
	require.Exactly(t, configDto.Approvers["two"], []string{"z", "o", "v"})
}

func TestExpandWatchers(t *testing.T) {

	instance := createInstance()

	testWatchers := []string{"x", "y", "z", "z"}

	result := instance.expandUserGroups(context.TODO(), testWatchers)

	require.Exactly(t, result, []string{"x", "y", "z"})
}

func createInstance() Impl {
	return createInstanceWithOwners(&owners.Impl{
		Configuration: nil,
		Logging:       nil,
		Cache:         nil,
		Updater:       nil,
	})
}

func createInstanceWithOwners(ownersImpl service.Owners) Impl {
	instance := Impl{
		Configuration: nil,
		Logging:       nil,
		Cache:         nil,
		Updater:       nil,
		Owners:        ownersImpl,
	}
	return instance
}

func createRepositoryConfigDto(testApprovers map[string][]string) *openapi.RepositoryConfigurationDto {
	return &openapi.RepositoryConfigurationDto{
		AccessKeys:              nil,
		CommitMessageType:       nil,
		RequireIssue:            nil,
		RequireSuccessfulBuilds: nil,
		Webhooks:                nil,
		Approvers:               testApprovers,
		Watchers:                nil,
	}
}
