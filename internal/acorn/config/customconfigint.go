package config

import (
	"regexp"

	"github.com/Interhyp/metadata-service/internal/types"
	"github.com/Roshick/go-autumn-kafka/pkg/kafka"

	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
)

type CustomConfiguration interface {
	BasicAuthUsername() string
	BasicAuthPassword() string

	ReviewerFallback() string
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

	AlertTargetRegex() *regexp.Regexp

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

	WebhooksProcessAsync() bool

	Kafka() *kafka.Config
	KafkaGroupIdOverride() string

	RedisUrl() string
	RedisPassword() string

	PullRequestBuildUrl() string
	PullRequestBuildKey() string

	GithubAppId() int64
	GithubAppInstallationId() int64
	GithubAppJwtSigningKeyPEM() []byte
	GithubAppWebhookSecret() []byte

	YamlIndentation() int
	FormattingActionCommitMsgPrefix() string

	CheckWarnMissingMainlineProtection() bool
	CheckExpectedRequiredConditions() []CheckedRequiredConditions
	CheckedExpectedExemptions() []CheckedExpectedExemption
}
type CheckedRequiredConditions struct {
	Name            string `yaml:"name" json:"name"`
	AnnotationLevel string `yaml:"annotationLevel" json:"annotationLevel"`
	RefMatcher      string `yaml:"refMatcher" json:"refMatcher"`
}

type CheckedExpectedExemption struct {
	Name       string   `yaml:"name" json:"name"`
	RefMatcher string   `yaml:"refMatcher" json:"refMatcher"`
	Exemptions []string `yaml:"exemptions" json:"exemptions"`
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
	KeyBasicAuthUsername                  = "BASIC_AUTH_USERNAME"
	KeyBasicAuthPassword                  = "BASIC_AUTH_PASSWORD"
	KeyReviewerFallback                   = "REVIEWER_FALLBACK"
	KeyGitCommitterName                   = "GIT_COMMITTER_NAME"
	KeyGitCommitterEmail                  = "GIT_COMMITTER_EMAIL"
	KeyKafkaGroupIdOverride               = "KAFKA_GROUP_ID_OVERRIDE"
	KeyAuthOidcKeySetUrl                  = "AUTH_OIDC_KEY_SET_URL"
	KeyAuthOidcTokenAudience              = "AUTH_OIDC_TOKEN_AUDIENCE"
	KeyAuthGroupWrite                     = "AUTH_GROUP_WRITE"
	KeyMetadataRepoUrl                    = "METADATA_REPO_URL"
	KeyMetadataRepoMainline               = "METADATA_REPO_MAINLINE"
	KeyUpdateJobIntervalMinutes           = "UPDATE_JOB_INTERVAL_MINUTES"
	KeyUpdateJobTimeoutSeconds            = "UPDATE_JOB_TIMEOUT_SECONDS"
	KeyAlertTargetRegex                   = "ALERT_TARGET_REGEX"
	KeyElasticApmDisabled                 = "ELASTIC_APM_DISABLED"
	KeyOwnerAliasPermittedRegex           = "OWNER_ALIAS_PERMITTED_REGEX"
	KeyOwnerAliasProhibitedRegex          = "OWNER_ALIAS_PROHIBITED_REGEX"
	KeyOwnerAliasMaxLength                = "OWNER_ALIAS_MAX_LENGTH"
	KeyOwnerAliasFilterRegex              = "OWNER_ALIAS_FILTER_REGEX"
	KeyServiceNamePermittedRegex          = "SERVICE_NAME_PERMITTED_REGEX"
	KeyServiceNameProhibitedRegex         = "SERVICE_NAME_PROHIBITED_REGEX"
	KeyServiceNameMaxLength               = "SERVICE_NAME_MAX_LENGTH"
	KeyRepositoryNamePermittedRegex       = "REPOSITORY_NAME_PERMITTED_REGEX"
	KeyRepositoryNameProhibitedRegex      = "REPOSITORY_NAME_PROHIBITED_REGEX"
	KeyRepositoryNameMaxLength            = "REPOSITORY_NAME_MAX_LENGTH"
	KeyRepositoryKeySeparator             = "REPOSITORY_KEY_SEPARATOR"
	KeyRepositoryTypes                    = "REPOSITORY_TYPES"
	KeyNotificationConsumerConfigs        = "NOTIFICATION_CONSUMER_CONFIGS"
	KeyRedisUrl                           = "REDIS_URL"
	KeyRedisPassword                      = "REDIS_PASSWORD"
	KeyPullRequestBuildUrl                = "PULL_REQUEST_BUILD_URL"
	KeyPullRequestBuildKey                = "PULL_REQUEST_BUILD_KEY"
	KeyWebhooksProcessAsync               = "WEBHOOKS_PROCESS_ASYNC"
	KeyGithubAppId                        = "GITHUB_APP_ID"
	KeyGithubAppInstallationId            = "GITHUB_APP_INSTALLATION_ID"
	KeyGithubAppJwtSigningKeyPEM          = "GITHUB_APP_JWT_SIGNING_KEY_PEM"
	KeyGithubAppWebhookSecret             = "GITHUB_APP_WEBHOOK_SECRET"
	KeyYamlIndentation                    = "YAML_INDENTATION"
	KeyFormattingActionCommitMsgPrefix    = "FORMATTING_ACTION_COMMIT_MSG_PREFIX"
	KeyCheckWarnMissingMainlineProtection = "CHECK_WARN_MISSING_MAINLINE_PROTECTION"
	KeyCheckExpectedRequiredConditions    = "CHECK_EXPECTED_REQUIRED_CONDITIONS"
	KeyCheckExpectedExemptions            = "CHECK_EXPECTED_EXEMPTIONS"
)
