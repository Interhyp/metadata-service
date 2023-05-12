package services

import (
	"context"
	"testing"
	"time"

	openapi "github.com/Interhyp/metadata-service/api/v1"
	"github.com/Interhyp/metadata-service/docs"
	"github.com/Interhyp/metadata-service/test/acceptance/configmock"
	auloggingapi "github.com/StephanHCB/go-autumn-logging/api"
	"github.com/StephanHCB/go-backend-service-common/api/apierrors"
	"github.com/StephanHCB/go-backend-service-common/repository/timestamp"
	"github.com/stretchr/testify/require"
)

func p(v string) *string {
	return &v
}

func b(v bool) *bool {
	return &v
}

func tstCurrent() openapi.ServiceDto {
	return openapi.ServiceDto{
		Owner: "owner",
		Quicklinks: []openapi.Quicklink{
			{
				Url:         p("url"),
				Title:       p("title"),
				Description: p("desc"),
			},
		},
		Repositories:    []string{"repo1", "repo2"},
		AlertTarget:     "target",
		DevelopmentOnly: b(true),
		OperationType:   p("PLATFORM"),
		RequiredScans:   []string{"SAST", "SCA"},
		TimeStamp:       "ts",
		CommitHash:      "hash",
	}
}

func tstPatchService(t *testing.T, patch openapi.ServicePatchDto, expected openapi.ServiceDto) {
	actual := patchService(tstCurrent(), patch)
	require.Equal(t, expected, actual)
}

func TestPatchService_EmptyPatch(t *testing.T) {
	docs.Description("patching of services works with an empty patch")
	expected := tstCurrent()
	expected.TimeStamp = "newts"
	expected.CommitHash = "newhash"
	tstPatchService(t, openapi.ServicePatchDto{
		TimeStamp:  "newts",
		CommitHash: "newhash",
	}, expected)
}

func TestPatchService_ReplaceAll(t *testing.T) {
	docs.Description("patching of services works with an all fields patch")
	tstPatchService(t, openapi.ServicePatchDto{
		Owner: p("newowner"),
		Quicklinks: []openapi.Quicklink{
			{
				Url:         p("newurl"),
				Title:       p("newtitle"),
				Description: p("newdesc"),
			},
		},
		Repositories:    []string{"repo3"},
		AlertTarget:     p("newtarget"),
		DevelopmentOnly: b(false),
		OperationType:   p("WORKLOAD"),
		RequiredScans:   []string{"SAST"},
		TimeStamp:       "newts",
		CommitHash:      "newhash",
	}, openapi.ServiceDto{
		Owner: "newowner",
		Quicklinks: []openapi.Quicklink{
			{
				Url:         p("newurl"),
				Title:       p("newtitle"),
				Description: p("newdesc"),
			},
		},
		Repositories:    []string{"repo3"},
		AlertTarget:     "newtarget",
		DevelopmentOnly: b(false),
		OperationType:   p("WORKLOAD"),
		RequiredScans:   []string{"SAST"},
		TimeStamp:       "newts",
		CommitHash:      "newhash",
	})
}

func TestPatchService_ClearFields(t *testing.T) {
	docs.Description("patching of services works with a patch that clears fields (result would not validate)")
	tstPatchService(t, openapi.ServicePatchDto{
		Owner:         p(""),
		Description:   p(""),
		Quicklinks:    []openapi.Quicklink{},
		Repositories:  []string{},
		AlertTarget:   p(""),
		OperationType: p(""),
		RequiredScans: []string{},
		TimeStamp:     "",
		CommitHash:    "",
	}, openapi.ServiceDto{
		Owner:           "",
		Description:     nil,
		AlertTarget:     "",
		DevelopmentOnly: b(true),
		OperationType:   nil,
		TimeStamp:       "",
		CommitHash:      "",
	})
}

func tstValid() openapi.ServiceDto {
	description := "short service description"
	return openapi.ServiceDto{
		Owner:       "some-owner",
		Description: &description,
		AlertTarget: "somebody@some-organisation.com",
		TimeStamp:   "timestamp",
		CommitHash:  "commithash",
		JiraIssue:   "jiraissue",
	}
}

func tstCreateValid() openapi.ServiceCreateDto {
	description := "short service description"
	return openapi.ServiceCreateDto{
		Owner:       "some-owner",
		Description: &description,
		AlertTarget: "somebody@some-organisation.com",
		JiraIssue:   "jiraissue",
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

func tstValidationTestcaseAllOps(t *testing.T, expectedMessage string, data openapi.ServiceDto, create openapi.ServiceCreateDto, patch openapi.ServicePatchDto) {
	mockConfig := configmock.MockConfig{}
	mockLogging := MockLogging{}
	fakeNow := func() time.Time {
		return time.Date(2022, 11, 6, 18, 14, 10, 0, time.UTC)
	}
	timestampImpl := timestamp.TimestampImpl{
		Timestamp: fakeNow,
	}
	impl := &Impl{
		Configuration:       nil,
		CustomConfiguration: &mockConfig,
		Logging:             &mockLogging,
		Cache:               nil,
		Updater:             nil,
		Timestamp:           &timestampImpl,
	}

	err := (*Impl).validateNewServiceDto(impl, context.TODO(), "any", create)
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, *err.(apierrors.AnnotatedError).ApiError().Details)

	err = (*Impl).validateExistingServiceDto(impl, context.TODO(), "any", data)
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, *err.(apierrors.AnnotatedError).ApiError().Details)

	patch.TimeStamp = "newts"
	patch.CommitHash = "newhash"
	patch.JiraIssue = "newjiraissue"

	err = (*Impl).validateServicePatchDto(impl, context.TODO(), "any", patch, data)
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, *err.(apierrors.AnnotatedError).ApiError().Details)
}

func TestValidate_AlertTarget(t *testing.T) {
	docs.Description("invalid alert targets are correctly rejected on all operations")

	data := tstValid()
	data.AlertTarget = "somethingelse"

	create := tstCreateValid()
	create.AlertTarget = "somethingelse"

	patch := openapi.ServicePatchDto{
		AlertTarget: p("somethingother"),
	}

	expectedMessage := "validation error: field alertTarget must either be an email address @some-organisation.com or a Teams webhook"

	tstValidationTestcaseAllOps(t, expectedMessage, data, create, patch)
}

func TestValidate_OperationType(t *testing.T) {
	docs.Description("invalid operation types are correctly rejected on all operations")

	data := tstValid()
	data.OperationType = p("OTHER")

	create := tstCreateValid()
	create.OperationType = p("OTHER")

	patch := openapi.ServicePatchDto{
		OperationType: p("YET ANOTHER"),
	}

	expectedMessage := "validation error: optional field operationType must be WORKLOAD (default if unset), PLATFORM or APPLICATION"

	tstValidationTestcaseAllOps(t, expectedMessage, data, create, patch)
}

func TestValidate_RequiredScans(t *testing.T) {
	docs.Description("invalid required scan types are correctly rejected on all operations")

	data := tstValid()
	data.RequiredScans = []string{"LASER"}

	create := tstCreateValid()
	create.RequiredScans = []string{"LASER"}

	patch := openapi.ServicePatchDto{
		RequiredScans: []string{"CUBIC"},
	}

	expectedMessage := "validation error: field requiredScans can only contain SAST and SCA"

	tstValidationTestcaseAllOps(t, expectedMessage, data, create, patch)
}
