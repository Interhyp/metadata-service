package config

import (
	auconfigapi "github.com/StephanHCB/go-autumn-config-api"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
)

const (
	KeyBbUser                   = "BB_USER"
	KeyGitCommitterName         = "GIT_COMMITTER_NAME"
	KeyGitCommitterEmail        = "GIT_COMMITTER_EMAIL"
	KeyKafkaUser                = "KAFKA_USER"
	KeyKafkaTopic               = "KAFKA_TOPIC"
	KeyKafkaSeedBrokers         = "KAFKA_SEED_BROKERS"
	KeyKafkaGroupIdOverride     = "KAFKA_GROUP_ID_OVERRIDE"
	KeyKeySetUrl                = "KEY_SET_URL"
	KeyMetadataRepoUrl          = "METADATA_REPO_URL"
	KeyOwnerRegex               = "OWNER_REGEX"
	KeyUpdateJobIntervalMinutes = "UPDATE_JOB_INTERVAL_MINUTES"
	KeyUpdateJobTimeoutSeconds  = "UPDATE_JOB_TIMEOUT_SECONDS"
	KeyVaultSecretsBasePath     = "VAULT_SECRETS_BASE_PATH"
	KeyVaultKafkaSecretPath     = "VAULT_KAFKA_SECRET_PATH"
	KeyAlertTargetPrefix        = "ALERT_TARGET_PREFIX"
	KeyAlertTargetSuffix        = "ALERT_TARGET_SUFFIX"
	KeyAdditionalPromoters      = "ADDITIONAL_PROMOTERS_FROM_OWNERS"
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
		Key:         KeyOwnerRegex,
		EnvName:     KeyOwnerRegex,
		Default:     ".*",
		Description: "regular expression to filter owners. Useful on localhost or for test instances to speed up service startup.",
		Validate:    auconfigapi.ConfigNeedsNoValidation,
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
		Description: "owner aliases from which to get additional promoters to be added for all services. Can be left empty, or contain a comma separated list of owner aliases",
		Validate:    auconfigenv.ObtainPatternValidator("^|[a-z](-?[a-z0-9]+)*(,[a-z](-?[a-z0-9]+)*)*$"),
	},
}
