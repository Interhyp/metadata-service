package repository

import (
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"regexp"
)

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

	UpdateJobIntervalCronPart() string
	UpdateJobTimeoutSeconds() uint16

	VaultSecretsBasePath() string
	VaultKafkaSecretPath() string

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
