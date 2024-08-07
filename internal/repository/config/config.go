package config

import (
	"math"

	"github.com/Interhyp/metadata-service/internal/acorn/config"
	auconfigapi "github.com/StephanHCB/go-autumn-config-api"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
)

var CustomConfigItems = []auconfigapi.ConfigItem{
	{
		Key:         config.KeyBasicAuthUsername,
		EnvName:     config.KeyBasicAuthUsername,
		Default:     "",
		Description: "username for basic-auth write access to this service",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         config.KeyBasicAuthPassword,
		EnvName:     config.KeyBasicAuthPassword,
		Default:     "",
		Description: "password for basic-auth write access to this service",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         config.KeySSHPrivateKey,
		EnvName:     config.KeySSHPrivateKey,
		Description: "ssh private key used to access the vcs",
		Default:     "",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         config.KeySSHPrivateKeyPassword,
		EnvName:     config.KeySSHPrivateKeyPassword,
		Description: "ssh private key password",
		Default:     "",
		Validate:    auconfigapi.ConfigNeedsNoValidation,
	},
	{
		Key:         config.KeySSHMetadataRepositoryUrl,
		EnvName:     config.KeySSHMetadataRepositoryUrl,
		Default:     "",
		Description: "ssh git clone url for service-metadata repository",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         config.KeyBitbucketUsername,
		EnvName:     config.KeyBitbucketUsername,
		Default:     "",
		Description: "bitbucket username for api and git clone service-metadata access",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         config.KeyBitbucketPassword,
		EnvName:     config.KeyBitbucketPassword,
		Default:     "",
		Description: "bitbucket password for api and git clone service-metadata access",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         config.KeyBitbucketServer,
		EnvName:     config.KeyBitbucketServer,
		Default:     "https://bitbucket.com",
		Description: "base URL to the bitbucket server, including protocol",
		Validate:    auconfigenv.ObtainPatternValidator("^https?://.*$"),
	},
	{
		Key:         config.KeyBitbucketCacheSize,
		EnvName:     config.KeyBitbucketCacheSize,
		Default:     "1000",
		Description: "size of the cache for bitbucket api items",
		Validate:    auconfigenv.ObtainUintRangeValidator(10, 1000),
	},
	{
		Key:         config.KeyBitbucketCacheRetentionSeconds,
		EnvName:     config.KeyBitbucketCacheRetentionSeconds,
		Default:     "3600",
		Description: "seconds to keep items in the bitbucket api cache (only used for bitbucket users)",
		Validate:    auconfigenv.ObtainUintRangeValidator(60, math.MaxUint32),
	},
	{
		Key:         config.KeyBitbucketReviewerFallback,
		EnvName:     config.KeyBitbucketReviewerFallback,
		Default:     "",
		Description: "default fallback reviewer username or groupname",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         config.KeyGitCommitterName,
		EnvName:     config.KeyGitCommitterName,
		Default:     "",
		Description: "name to use for git commits",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         config.KeyGitCommitterEmail,
		EnvName:     config.KeyGitCommitterEmail,
		Default:     "",
		Description: "email address to use for git commits",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         config.KeyKafkaGroupIdOverride,
		EnvName:     config.KeyKafkaGroupIdOverride,
		Default:     "",
		Description: "optional: a kafka group id to use for subscribing to update events. Mainly useful on localhost. If empty, group id is derived from 3rd oktet of non-trivial local ip (as proxy for the k8s worker node)",
		Validate:    auconfigenv.ObtainPatternValidator("^(|[a-z0-9-]+)$"),
	},
	{
		Key:         config.KeyAuthOidcKeySetUrl,
		EnvName:     config.KeyAuthOidcKeySetUrl,
		Default:     "",
		Description: "keyset url of oidc identity provider",
		Validate:    auconfigenv.ObtainPatternValidator("^https?:.*$"),
	},
	{
		Key:         config.KeyAuthOidcTokenAudience,
		EnvName:     config.KeyAuthOidcTokenAudience,
		Default:     "",
		Description: "expected audience of oidc access token",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         config.KeyAuthGroupWrite,
		EnvName:     config.KeyAuthGroupWrite,
		Default:     "",
		Description: "group name or id for write access to this service",
		Validate:    auconfigapi.ConfigNeedsNoValidation,
	},
	{
		Key:         config.KeyMetadataRepoUrl,
		EnvName:     config.KeyMetadataRepoUrl,
		Default:     "",
		Description: "git clone url for service-metadata repository",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         config.KeyMetadataRepoMainline,
		EnvName:     config.KeyMetadataRepoMainline,
		Default:     "refs/heads/main",
		Description: "ref to use as mainline",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         config.KeyUpdateJobIntervalMinutes,
		EnvName:     config.KeyUpdateJobIntervalMinutes,
		Default:     "5",
		Description: "time in minutes between cache update. Must be a divisor of 60 (used in cron expression) - pick one of the choices",
		Validate:    auconfigenv.ObtainPatternValidator("^(1|2|3|4|5|6|10|12|15|20|30)$"),
	},
	{
		Key:         config.KeyUpdateJobTimeoutSeconds,
		EnvName:     config.KeyUpdateJobTimeoutSeconds,
		Default:     "30",
		Description: "timeout for the cache update job in seconds. Must be less than 60 * UPDATE_JOB_INTERVAL_MINUTES",
		Validate:    auconfigenv.ObtainUintRangeValidator(10, 60),
	},
	{
		Key:      config.KeyAlertTargetRegex,
		EnvName:  config.KeyAlertTargetRegex,
		Default:  "",
		Validate: auconfigenv.ObtainIsRegexValidator(),
	},
	{
		Key:         config.KeyElasticApmDisabled,
		EnvName:     config.KeyElasticApmDisabled,
		Default:     "false",
		Description: "disable elastic apm middleware. supports all values supported by ParseBool (https://pkg.go.dev/strconv#ParseBool).",
		Validate:    auconfigenv.ObtainIsBooleanValidator(),
	},
	{
		Key:         config.KeyOwnerAliasPermittedRegex,
		EnvName:     config.KeyOwnerAliasPermittedRegex,
		Default:     "^[a-z](-?[a-z0-9]+)*$",
		Description: "regular expression to control the owner aliases that are permitted to be be created.",
		Validate:    auconfigenv.ObtainIsRegexValidator(),
	},
	{
		Key:         config.KeyOwnerAliasProhibitedRegex,
		EnvName:     config.KeyOwnerAliasProhibitedRegex,
		Default:     "^$",
		Description: "regular expression to control the owner aliases that are prohibited to be be created.",
		Validate:    auconfigenv.ObtainIsRegexValidator(),
	},
	{
		Key:         config.KeyOwnerAliasFilterRegex,
		EnvName:     config.KeyOwnerAliasFilterRegex,
		Default:     "^.*$",
		Description: "regular expression to filter owners based on their alias. Useful on localhost or for test instances to speed up service startup.",
		Validate:    auconfigenv.ObtainIsRegexValidator(),
	},
	{
		Key:         config.KeyOwnerAliasMaxLength,
		EnvName:     config.KeyOwnerAliasMaxLength,
		Default:     "28",
		Description: "maximum length of a valid owner alias.",
		Validate:    auconfigenv.ObtainIntRangeValidator(1, 100),
	},
	{
		Key:         config.KeyServiceNamePermittedRegex,
		EnvName:     config.KeyServiceNamePermittedRegex,
		Default:     "^[a-z](-?[a-z0-9]+)*$",
		Description: "regular expression to control the service names that are permitted to be be created.",
		Validate:    auconfigenv.ObtainIsRegexValidator(),
	},
	{
		Key:         config.KeyServiceNameProhibitedRegex,
		EnvName:     config.KeyServiceNameProhibitedRegex,
		Default:     "^$",
		Description: "regular expression to control the service names that are prohibited to be be created.",
		Validate:    auconfigenv.ObtainIsRegexValidator(),
	},
	{
		Key:         config.KeyServiceNameMaxLength,
		EnvName:     config.KeyServiceNameMaxLength,
		Default:     "28",
		Description: "maximum length of a valid service name.",
		Validate:    auconfigenv.ObtainIntRangeValidator(1, 100),
	},
	{
		Key:         config.KeyRepositoryNamePermittedRegex,
		EnvName:     config.KeyRepositoryNamePermittedRegex,
		Default:     "^[a-z](-?[a-z0-9]+)*$",
		Description: "regular expression to control the repository names that are permitted to be be created.",
		Validate:    auconfigenv.ObtainIsRegexValidator(),
	},
	{
		Key:         config.KeyRepositoryNameProhibitedRegex,
		EnvName:     config.KeyRepositoryNameProhibitedRegex,
		Default:     "^$",
		Description: "regular expression to control the repository names that are prohibited to be be created.",
		Validate:    auconfigenv.ObtainIsRegexValidator(),
	},
	{
		Key:         config.KeyRepositoryNameMaxLength,
		EnvName:     config.KeyRepositoryNameMaxLength,
		Default:     "64",
		Description: "maximum length of a valid repository name.",
		Validate:    auconfigenv.ObtainIntRangeValidator(1, 100),
	},
	{
		Key:         config.KeyRepositoryTypes,
		EnvName:     config.KeyRepositoryTypes,
		Default:     "",
		Description: "comma separated list of supported repository types.",
		Validate:    auconfigenv.ObtainPatternValidator("^|[a-z](-?[a-z0-9]+)*(,[a-z](-?[a-z0-9]+)*)*$"),
	},
	{
		Key:         config.KeyRepositoryKeySeparator,
		EnvName:     config.KeyRepositoryKeySeparator,
		Default:     ".",
		Description: "single character used to separate repository name from repository type. repository name and repository type must not contain separator.",
		Validate:    auconfigenv.ObtainSingleCharacterValidator(),
	},
	{
		Key:         config.KeyNotificationConsumerConfigs,
		EnvName:     config.KeyNotificationConsumerConfigs,
		Default:     "",
		Description: "configurations for consumers of change notifications.",
		Validate: func(key string) error {
			value := auconfigenv.Get(key)
			_, err := parseNotificationConsumerConfigs(value)
			return err
		},
	},
	{
		Key:         config.KeyAllowedFileCategories,
		EnvName:     config.KeyAllowedFileCategories,
		Default:     "",
		Description: "allowed filecategory keys",
		Validate: func(key string) error {
			value := auconfigenv.Get(key)
			_, err := parseAllowedFileCategories(value)
			return err
		},
	},
	{
		Key:         config.KeyRedisUrl,
		EnvName:     config.KeyRedisUrl,
		Default:     "",
		Description: "base url to the redis, including protocol. Uses in-memory caching if blank.",
		Validate:    auconfigapi.ConfigNeedsNoValidation,
	},
	{
		Key:         config.KeyRedisPassword,
		EnvName:     config.KeyRedisPassword,
		Default:     "",
		Description: "password used to access the redis",
		Validate:    auconfigapi.ConfigNeedsNoValidation,
	},
	{
		Key:         config.KeyPullRequestBuildUrl,
		EnvName:     config.KeyPullRequestBuildUrl,
		Default:     "",
		Description: "Url that pull request builds should link to.",
		Validate:    auconfigenv.ObtainPatternValidator("^https?://.*$"),
	},
	{
		Key:         config.KeyPullRequestBuildKey,
		EnvName:     config.KeyPullRequestBuildKey,
		Default:     "metadata-service",
		Description: "Key to use for pull request builds.",
		Validate:    auconfigapi.ConfigNeedsNoValidation,
	},
}
