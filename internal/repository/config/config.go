package config

import (
	"fmt"
	auconfigapi "github.com/StephanHCB/go-autumn-config-api"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
	"regexp"
	"strconv"
)

const (
	KeyBbUser                        = "BB_USER"
	KeyGitCommitterName              = "GIT_COMMITTER_NAME"
	KeyGitCommitterEmail             = "GIT_COMMITTER_EMAIL"
	KeyKafkaUser                     = "KAFKA_USER"
	KeyKafkaTopic                    = "KAFKA_TOPIC"
	KeyKafkaSeedBrokers              = "KAFKA_SEED_BROKERS"
	KeyKafkaGroupIdOverride          = "KAFKA_GROUP_ID_OVERRIDE"
	KeyKeySetUrl                     = "KEY_SET_URL"
	KeyMetadataRepoUrl               = "METADATA_REPO_URL"
	KeyUpdateJobIntervalMinutes      = "UPDATE_JOB_INTERVAL_MINUTES"
	KeyUpdateJobTimeoutSeconds       = "UPDATE_JOB_TIMEOUT_SECONDS"
	KeyVaultSecretsBasePath          = "VAULT_SECRETS_BASE_PATH"
	KeyVaultKafkaSecretPath          = "VAULT_KAFKA_SECRET_PATH"
	KeyAlertTargetPrefix             = "ALERT_TARGET_PREFIX"
	KeyAlertTargetSuffix             = "ALERT_TARGET_SUFFIX"
	KeyAdditionalPromoters           = "ADDITIONAL_PROMOTERS"
	KeyAdditionalPromotersFromOwners = "ADDITIONAL_PROMOTERS_FROM_OWNERS"
	KeyElasticApmDisabled            = "ELASTIC_APM_DISABLED"
	KeyOwnerPermittedAliasRegex      = "OWNER_PERMITTED_ALIAS_REGEX"
	KeyOwnerProhibitedAliasRegex     = "OWNER_PROHIBITED_ALIAS_REGEX"
	KeyOwnerFilterAliasRegex         = "OWNER_FILTER_ALIAS_REGEX"
	KeyServicePermittedNameRegex     = "SERVICE_PERMITTED_NAME_REGEX"
	KeyServiceProhibitedNameRegex    = "SERVICE_PROHIBITED_NAME_REGEX"
	KeyRepositoryPermittedNameRegex  = "REPOSITORY_PERMITTED_NAME_REGEX"
	KeyRepositoryProhibitedNameRegex = "REPOSITORY_PROHIBITED_NAME_REGEX"
	KeyRepositoryTypes               = "REPOSITORY_TYPES"
	KeyRepositoryKeySeparator        = "REPOSITORY_KEY_SEPARATOR"
)

var CustomConfigItems = []auconfigapi.ConfigItem{
	{
		Key:         KeyBbUser,
		EnvName:     KeyBbUser,
		Default:     "",
		Description: "bitbucket username for api and git clone service-metadata access",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         KeyGitCommitterName,
		EnvName:     KeyGitCommitterName,
		Default:     "",
		Description: "name to use for git commits",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         KeyGitCommitterEmail,
		EnvName:     KeyGitCommitterEmail,
		Default:     "",
		Description: "email address to use for git commits",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         KeyKafkaUser,
		EnvName:     KeyKafkaUser,
		Default:     "",
		Description: "optional: kafka user (needed to send kafka notifications), leaving this or any of the other *KAFKA* fields empty will switch off all Kafka functionality",
		Validate:    auconfigenv.ObtainPatternValidator("^(|[a-z0-9-]+)$"),
	},
	{
		Key:         KeyKafkaTopic,
		EnvName:     KeyKafkaTopic,
		Default:     "",
		Description: "optional: kafka user (needed to send kafka notifications), leaving this or any of the other *KAFKA* fields empty will switch off all Kafka functionality",
		Validate:    auconfigenv.ObtainPatternValidator("^(|[a-z0-9-]+)$"),
	},
	{
		Key:         KeyKafkaSeedBrokers,
		EnvName:     KeyKafkaSeedBrokers,
		Default:     "",
		Description: "optional: comma separated list of kafka seed broker URLs (needed to send kafka notifications), leaving this or any of the other *KAFKA* fields empty will switch off all Kafka functionality",
		Validate:    auconfigenv.ObtainPatternValidator("^(|([a-z0-9-]+.[a-z0-9-]+.[a-z]{2,3}:9092)(,[a-z0-9-]+.[a-z0-9-]+.[a-z]{2,3}:9092)*)$"),
	},
	{
		Key:         KeyKafkaGroupIdOverride,
		EnvName:     KeyKafkaGroupIdOverride,
		Default:     "",
		Description: "optional: a kafka group id to use for subscribing to update events. Mainly useful on localhost. If empty, group id is derived from 3rd oktet of non-trivial local ip (as proxy for the k8s worker node)",
		Validate:    auconfigenv.ObtainPatternValidator("^(|[a-z0-9-]+)$"),
	},
	{
		Key:         KeyKeySetUrl,
		EnvName:     KeyKeySetUrl,
		Default:     "",
		Description: "keyset URL of your OIDC identity provider",
		Validate:    auconfigenv.ObtainPatternValidator("^https?:.*$"),
	},
	{
		Key:         KeyMetadataRepoUrl,
		EnvName:     KeyMetadataRepoUrl,
		Default:     "",
		Description: "git clone url for service-metadata repository",
		Validate:    auconfigenv.ObtainNotEmptyValidator(),
	},
	{
		Key:         KeyUpdateJobIntervalMinutes,
		EnvName:     KeyUpdateJobIntervalMinutes,
		Default:     "5",
		Description: "time in minutes between cache update. Must be a divisor of 60 (used in cron expression) - pick one of the choices",
		Validate:    auconfigenv.ObtainPatternValidator("^(1|2|3|4|5|6|10|12|15|20|30)$"),
	},
	{
		Key:         KeyUpdateJobTimeoutSeconds,
		EnvName:     KeyUpdateJobTimeoutSeconds,
		Default:     "30",
		Description: "timeout for the cache update job in seconds. Must be less than 60 * UPDATE_JOB_INTERVAL_MINUTES",
		Validate:    auconfigenv.ObtainUintRangeValidator(10, 60),
	},
	{
		Key:         KeyVaultSecretsBasePath,
		EnvName:     KeyVaultSecretsBasePath,
		Default:     "",
		Description: "total vaul secret path is composed of VAULT_SECRETS_BASE_PATH/ENVIRONMENT/VAULT_SECRET_PATH",
		Validate:    auconfigenv.ObtainPatternValidator("^(|[a-z0-9-/]+)$"),
	},
	{
		Key:         KeyVaultKafkaSecretPath,
		EnvName:     KeyVaultKafkaSecretPath,
		Default:     "",
		Description: "optional: kafka secret path in vault (needed to send kafka notifications), leaving this or any of the other *KAFKA* fields empty will switch off all Kafka functionality, including the Vault query for Kafka credentials",
		Validate:    auconfigenv.ObtainPatternValidator("^(|[a-z0-9-/]+)$"),
	},
	{
		Key:      KeyAlertTargetPrefix,
		EnvName:  KeyAlertTargetPrefix,
		Default:  "",
		Validate: auconfigenv.ObtainPatternValidator("^((http|https)://|)[a-z0-9-.]+.[a-z]{2,3}/$"),
	},
	{
		Key:      KeyAlertTargetSuffix,
		EnvName:  KeyAlertTargetSuffix,
		Default:  "",
		Validate: auconfigenv.ObtainPatternValidator("^@[a-z0-9-]+.[a-z]{2,3}$"),
	},
	{
		Key:         KeyAdditionalPromoters,
		EnvName:     KeyAdditionalPromoters,
		Default:     "",
		Description: "promoters to be added for all services. Can be left empty, or contain a comma separated list of usernames",
		Validate:    auconfigenv.ObtainPatternValidator("^|[a-z](-?[a-z0-9]+)*(,[a-z](-?[a-z0-9]+)*)*$"),
	},
	{
		Key:         KeyAdditionalPromotersFromOwners,
		EnvName:     KeyAdditionalPromotersFromOwners,
		Default:     "",
		Description: "owner aliases from which to get additional promoters to be added for all services. Can be left empty, or contain a comma separated list of owner aliases",
		Validate:    auconfigenv.ObtainPatternValidator("^|[a-z](-?[a-z0-9]+)*(,[a-z](-?[a-z0-9]+)*)*$"),
	},
	{
		Key:         KeyElasticApmDisabled,
		EnvName:     KeyElasticApmDisabled,
		Default:     "false",
		Description: "Disable Elastic APM middleware. Supports all values supported by ParseBool (https://pkg.go.dev/strconv#ParseBool).",
		Validate:    booleanValidator,
	},
	{
		Key:         KeyOwnerPermittedAliasRegex,
		EnvName:     KeyOwnerPermittedAliasRegex,
		Default:     "^[a-z](-?[a-z0-9]+)*$",
		Description: "regular expression to control the owner aliases that are permitted to be be created.",
		Validate:    regexCompileValidator,
	},
	{
		Key:         KeyOwnerProhibitedAliasRegex,
		EnvName:     KeyOwnerProhibitedAliasRegex,
		Default:     "^$",
		Description: "regular expression to control the owner aliases that are prohibited to be be created.",
		Validate:    regexCompileValidator,
	},
	{
		Key:         KeyOwnerFilterAliasRegex,
		EnvName:     KeyOwnerFilterAliasRegex,
		Default:     "^.*$",
		Description: "regular expression to filter owners based on their alias. Useful on localhost or for test instances to speed up service startup.",
		Validate:    regexCompileValidator,
	},
	{
		Key:         KeyServicePermittedNameRegex,
		EnvName:     KeyServicePermittedNameRegex,
		Default:     "^[a-z](-?[a-z0-9]+)*$",
		Description: "regular expression to control the service names that are permitted to be be created.",
		Validate:    regexCompileValidator,
	},
	{
		Key:         KeyServiceProhibitedNameRegex,
		EnvName:     KeyServiceProhibitedNameRegex,
		Default:     "^$",
		Description: "regular expression to control the service names that are prohibited to be be created.",
		Validate:    regexCompileValidator,
	},
	{
		Key:         KeyRepositoryPermittedNameRegex,
		EnvName:     KeyRepositoryPermittedNameRegex,
		Default:     "^[a-z](-?[a-z0-9]+)*$",
		Description: "regular expression to control the repository names that are permitted to be be created.",
		Validate:    regexCompileValidator,
	},
	{
		Key:         KeyRepositoryProhibitedNameRegex,
		EnvName:     KeyRepositoryProhibitedNameRegex,
		Default:     "^$",
		Description: "regular expression to control the repository names that are prohibited to be be created.",
		Validate:    regexCompileValidator,
	},
	{
		Key:         KeyRepositoryTypes,
		EnvName:     KeyRepositoryTypes,
		Default:     "",
		Description: "comma separated list of supported repository types.",
		Validate:    auconfigenv.ObtainPatternValidator("^|[a-z](-?[a-z0-9]+)*(,[a-z](-?[a-z0-9]+)*)*$"),
	},
	{
		Key:         KeyRepositoryKeySeparator,
		EnvName:     KeyRepositoryKeySeparator,
		Default:     ".",
		Description: "single character used to separate repository name from repository type. repository name and repository type must not contain separator.",
		Validate:    singleCharacterValidator,
	},
}

func booleanValidator(key string) error {
	value := auconfigenv.Get(key)
	_, err := strconv.ParseBool(value)
	return err
}

func singleCharacterValidator(key string) error {
	value := auconfigenv.Get(key)
	if len(value) < 1 {
		return fmt.Errorf("parameter cannot be empty")
	} else if len(value) > 1 {
		return fmt.Errorf("parameter cannot consist of multiple characters")
	}
	return nil
}

func regexCompileValidator(key string) error {
	value := auconfigenv.Get(key)
	_, err := regexp.Compile(value)
	return err
}
