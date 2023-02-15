package services

import (
	"context"
	openapi "github.com/Interhyp/metadata-service/api/v1"
	"github.com/Interhyp/metadata-service/docs"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
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

type MockConfig struct {
}

func (c *MockConfig) BasicAuthUsername() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) BasicAuthPassword() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) BitbucketUsername() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) BitbucketPassword() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) BitbucketServer() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) BitbucketCacheSize() int {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) BitbucketCacheRetentionSeconds() uint32 {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) BitbucketReviewerFallback() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) GitCommitterName() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) GitCommitterEmail() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) KafkaUsername() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) KafkaPassword() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) KafkaTopic() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) KafkaSeedBrokers() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) KafkaGroupIdOverride() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) AuthOidcKeySetUrl() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) AuthOidcTokenAudience() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) AuthGroupWrite() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) MetadataRepoUrl() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) UpdateJobIntervalCronPart() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) UpdateJobTimeoutSeconds() uint16 {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) AlertTargetPrefix() string {
	return "https://some-domain.com/"
}

func (c *MockConfig) AlertTargetSuffix() string {
	return "@some-organisation.com"
}

func (c *MockConfig) AdditionalPromotersFromOwners() []string {
	return make([]string, 0)
}

func (c *MockConfig) AdditionalPromoters() []string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) ElasticApmEnabled() bool {
	return false
}

func (c *MockConfig) OwnerAliasPermittedRegex() *regexp.Regexp {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) OwnerAliasProhibitedRegex() *regexp.Regexp {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) OwnerAliasMaxLength() uint16 {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) OwnerFilterAliasRegex() *regexp.Regexp {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) ServiceNamePermittedRegex() *regexp.Regexp {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) ServiceNameProhibitedRegex() *regexp.Regexp {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) ServiceNameMaxLength() uint16 {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) RepositoryNamePermittedRegex() *regexp.Regexp {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) RepositoryNameProhibitedRegex() *regexp.Regexp {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) RepositoryNameMaxLength() uint16 {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) RepositoryTypes() []string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) RepositoryKeySeparator() string {
	//TODO implement me
	panic("implement me")
}

func tstValidationTestcaseAllOps(t *testing.T, expectedMessage string, data openapi.ServiceDto, create openapi.ServiceCreateDto, patch openapi.ServicePatchDto) {
	mockConfig := MockConfig{}
	impl := &Impl{
		Configuration:       nil,
		CustomConfiguration: &mockConfig,
		Logging:             nil,
		Cache:               nil,
		Updater:             nil,
	}

	err := (*Impl).validateNewServiceDto(impl, context.TODO(), "any", create)
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, err.Error())

	err = (*Impl).validateExistingServiceDto(impl, context.TODO(), "any", data)
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, err.Error())

	patch.TimeStamp = "newts"
	patch.CommitHash = "newhash"
	patch.JiraIssue = "newjiraissue"

	err = (*Impl).validateServicePatchDto(impl, context.TODO(), "any", patch, data)
	require.NotNil(t, err)
	require.Equal(t, expectedMessage, err.Error())
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

	expectedMessage := "validation error: optional field operationType must be WORKLOAD (default if unset) or PLATFORM"

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
