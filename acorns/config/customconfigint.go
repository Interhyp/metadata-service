package config

import (
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"regexp"
)

type CustomConfiguration interface {
	BasicAuthUsername() string
	BasicAuthPassword() string

	BitbucketUsername() string
	BitbucketPassword() string

	GitCommitterName() string
	GitCommitterEmail() string

	KafkaUsername() string
	KafkaPassword() string
	KafkaTopic() string
	KafkaSeedBrokers() string
	KafkaGroupIdOverride() string

	KeySetUrl() string

	MetadataRepoUrl() string

	UpdateJobIntervalCronPart() string
	UpdateJobTimeoutSeconds() uint16

	AlertTargetPrefix() string
	AlertTargetSuffix() string

	AdditionalPromoters() []string
	AdditionalPromotersFromOwners() []string

	ElasticApmEnabled() bool

	OwnerAliasPermittedRegex() *regexp.Regexp
	OwnerAliasProhibitedRegex() *regexp.Regexp
	OwnerAliasMaxLength() uint16
	OwnerFilterAliasRegex() *regexp.Regexp

	ServiceNamePermittedRegex() *regexp.Regexp
	ServiceNameProhibitedRegex() *regexp.Regexp
	ServiceNameMaxLength() uint16

	RepositoryNamePermittedRegex() *regexp.Regexp
	RepositoryNameProhibitedRegex() *regexp.Regexp
	RepositoryNameMaxLength() uint16
	RepositoryTypes() []string
	RepositoryKeySeparator() string
}

// Custom is a type casting helper that gets you from the configuration acorn to your CustomConfiguration
func Custom(configuration librepo.Configuration) CustomConfiguration {
	return configuration.Custom().(CustomConfiguration)
}

const (
	KeyBasicAuthUsername             = "BASIC_AUTH_USERNAME"
	KeyBasicAuthPassword             = "BASIC_AUTH_PASSWORD"
	KeyBitbucketUsername             = "BITBUCKET_USERNAME"
	KeyBitbucketPassword             = "BITBUCKET_PASSWORD"
	KeyGitCommitterName              = "GIT_COMMITTER_NAME"
	KeyGitCommitterEmail             = "GIT_COMMITTER_EMAIL"
	KeyKafkaUsername                 = "KAFKA_USERNAME"
	KeyKafkaPassword                 = "KAFKA_PASSWORD"
	KeyKafkaTopic                    = "KAFKA_TOPIC"
	KeyKafkaSeedBrokers              = "KAFKA_SEED_BROKERS"
	KeyKafkaGroupIdOverride          = "KAFKA_GROUP_ID_OVERRIDE"
	KeyKeySetUrl                     = "KEY_SET_URL"
	KeyMetadataRepoUrl               = "METADATA_REPO_URL"
	KeyUpdateJobIntervalMinutes      = "UPDATE_JOB_INTERVAL_MINUTES"
	KeyUpdateJobTimeoutSeconds       = "UPDATE_JOB_TIMEOUT_SECONDS"
	KeyAlertTargetPrefix             = "ALERT_TARGET_PREFIX"
	KeyAlertTargetSuffix             = "ALERT_TARGET_SUFFIX"
	KeyAdditionalPromoters           = "ADDITIONAL_PROMOTERS"
	KeyAdditionalPromotersFromOwners = "ADDITIONAL_PROMOTERS_FROM_OWNERS"
	KeyElasticApmDisabled            = "ELASTIC_APM_DISABLED"
	KeyOwnerAliasPermittedRegex      = "OWNER_ALIAS_PERMITTED_REGEX"
	KeyOwnerAliasProhibitedRegex     = "OWNER_ALIAS_PROHIBITED_REGEX"
	KeyOwnerAliasMaxLength           = "OWNER_ALIAS_MAX_LENGTH"
	KeyOwnerAliasFilterRegex         = "OWNER_ALIAS_FILTER_REGEX"
	KeyServiceNamePermittedRegex     = "SERVICE_NAME_PERMITTED_REGEX"
	KeyServiceNameProhibitedRegex    = "SERVICE_NAME_PROHIBITED_REGEX"
	KeyServiceNameMaxLength          = "SERVICE_NAME_MAX_LENGTH"
	KeyRepositoryNamePermittedRegex  = "REPOSITORY_NAME_PERMITTED_REGEX"
	KeyRepositoryNameProhibitedRegex = "REPOSITORY_NAME_PROHIBITED_REGEX"
	KeyRepositoryNameMaxLength       = "REPOSITORY_NAME_MAX_LENGTH"
	KeyRepositoryKeySeparator        = "REPOSITORY_KEY_SEPARATOR"
	KeyRepositoryTypes               = "REPOSITORY_TYPES"
)
