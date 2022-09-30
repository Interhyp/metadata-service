package repository

import librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"

type CustomConfiguration interface {
	BbUser() string

	GitCommitterName() string
	GitCommitterEmail() string

	KafkaUser() string
	KafkaTopic() string
	KafkaSeedBrokers() string
	KafkaGroupIdOverride() string

	KeySetUrl() string

	MetadataRepoUrl() string

	OwnerRegex() string

	UpdateJobIntervalCronPart() string
	UpdateJobTimeoutSeconds() uint16

	VaultSecretsBasePath() string
	VaultKafkaSecretPath() string

	AlertTargetPrefix() string
	AlertTargetSuffix() string
}

// Custom is a type casting helper that gets you from the configuration acorn to your CustomConfiguration
func Custom(configuration librepo.Configuration) CustomConfiguration {
	return configuration.Custom().(CustomConfiguration)
}
