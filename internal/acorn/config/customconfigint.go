package config

import (
	"github.com/Interhyp/metadata-service/internal/types"
	"github.com/Roshick/go-autumn-kafka/pkg/aukafka"
	"regexp"

	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
)

type CustomConfiguration interface {
	BasicAuthUsername() string
	BasicAuthPassword() string

	BitbucketUsername() string
	BitbucketPassword() string

	SSHPrivateKey() string
	SSHPrivateKeyPassword() string
	SSHMetadataRepositoryUrl() string

	BitbucketServer() string
	BitbucketCacheSize() int
	BitbucketCacheRetentionSeconds() uint32
	BitbucketReviewerFallback() string

	GitCommitterName() string
	GitCommitterEmail() string

	AuthOidcKeySetUrl() string
	AuthOidcTokenAudience() string
	AuthGroupWrite() string

	MetadataRepoUrl() string
	MetadataRepoMainline() string
	MetadataRepoProject() string
	MetadataRepoName() string

	UpdateJobIntervalCronPart() string
	UpdateJobTimeoutSeconds() uint16

	AlertTargetPrefix() string
	AlertTargetSuffix() string

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

	NotificationConsumerConfigs() map[string]NotificationConsumerConfig

	AllowedFileCategories() []string

	Kafka() *aukafka.Config
	KafkaGroupIdOverride() string

	RedisUrl() string
	RedisPassword() string

	PullRequestBuildUrl() string
	PullRequestBuildKey() string
}

type NotificationConsumerConfig struct {
	Subscribed  map[types.NotificationPayloadType]map[types.NotificationEventType]struct{}
	ConsumerURL string
}

// Custom is a type casting helper that gets you from the configuration acorn to your CustomConfiguration
func Custom(configuration librepo.Configuration) CustomConfiguration {
	return configuration.Custom().(CustomConfiguration)
}

const (
	KeyBasicAuthUsername              = "BASIC_AUTH_USERNAME"
	KeyBasicAuthPassword              = "BASIC_AUTH_PASSWORD"
	KeySSHPrivateKey                  = "SSH_PRIVATE_KEY"
	KeySSHPrivateKeyPassword          = "SSH_PRIVATE_KEY_PASSWORD"
	KeySSHMetadataRepositoryUrl       = "SSH_METADATA_REPO_URL"
	KeyBitbucketUsername              = "BITBUCKET_USERNAME"
	KeyBitbucketPassword              = "BITBUCKET_PASSWORD"
	KeyBitbucketServer                = "BITBUCKET_SERVER"
	KeyBitbucketCacheSize             = "BITBUCKET_CACHE_SIZE"
	KeyBitbucketCacheRetentionSeconds = "BITBUCKET_CACHE_RETENTION_SECONDS"
	KeyBitbucketReviewerFallback      = "BITBUCKET_REVIEWER_FALLBACK"
	KeyGitCommitterName               = "GIT_COMMITTER_NAME"
	KeyGitCommitterEmail              = "GIT_COMMITTER_EMAIL"
	KeyKafkaUsername                  = "KAFKA_USERNAME"
	KeyKafkaPassword                  = "KAFKA_PASSWORD"
	KeyKafkaTopic                     = "KAFKA_TOPIC"
	KeyKafkaSeedBrokers               = "KAFKA_SEED_BROKERS"
	KeyKafkaGroupIdOverride           = "KAFKA_GROUP_ID_OVERRIDE"
	KeyAuthOidcKeySetUrl              = "AUTH_OIDC_KEY_SET_URL"
	KeyAuthOidcTokenAudience          = "AUTH_OIDC_TOKEN_AUDIENCE"
	KeyAuthGroupWrite                 = "AUTH_GROUP_WRITE"
	KeyMetadataRepoUrl                = "METADATA_REPO_URL"
	KeyMetadataRepoMainline           = "METADATA_REPO_MAINLINE"
	KeyUpdateJobIntervalMinutes       = "UPDATE_JOB_INTERVAL_MINUTES"
	KeyUpdateJobTimeoutSeconds        = "UPDATE_JOB_TIMEOUT_SECONDS"
	KeyAlertTargetPrefix              = "ALERT_TARGET_PREFIX"
	KeyAlertTargetSuffix              = "ALERT_TARGET_SUFFIX"
	KeyElasticApmDisabled             = "ELASTIC_APM_DISABLED"
	KeyOwnerAliasPermittedRegex       = "OWNER_ALIAS_PERMITTED_REGEX"
	KeyOwnerAliasProhibitedRegex      = "OWNER_ALIAS_PROHIBITED_REGEX"
	KeyOwnerAliasMaxLength            = "OWNER_ALIAS_MAX_LENGTH"
	KeyOwnerAliasFilterRegex          = "OWNER_ALIAS_FILTER_REGEX"
	KeyServiceNamePermittedRegex      = "SERVICE_NAME_PERMITTED_REGEX"
	KeyServiceNameProhibitedRegex     = "SERVICE_NAME_PROHIBITED_REGEX"
	KeyServiceNameMaxLength           = "SERVICE_NAME_MAX_LENGTH"
	KeyRepositoryNamePermittedRegex   = "REPOSITORY_NAME_PERMITTED_REGEX"
	KeyRepositoryNameProhibitedRegex  = "REPOSITORY_NAME_PROHIBITED_REGEX"
	KeyRepositoryNameMaxLength        = "REPOSITORY_NAME_MAX_LENGTH"
	KeyRepositoryKeySeparator         = "REPOSITORY_KEY_SEPARATOR"
	KeyRepositoryTypes                = "REPOSITORY_TYPES"
	KeyNotificationConsumerConfigs    = "NOTIFICATION_CONSUMER_CONFIGS"
	KeyAllowedFileCategories          = "ALLOWED_FILE_CATEGORIES"
	KeyRedisUrl                       = "REDIS_URL"
	KeyRedisPassword                  = "REDIS_PASSWORD"
	KeyPullRequestBuildUrl            = "PULL_REQUEST_BUILD_URL"
	KeyPullRequestBuildKey            = "PULL_REQUEST_BUILD_KEY"
)
