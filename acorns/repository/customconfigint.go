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

	OwnerPermittedAliasRegex() *regexp.Regexp
	OwnerProhibitedAliasRegex() *regexp.Regexp
	OwnerFilterAliasRegex() *regexp.Regexp

	ServicePermittedNameRegex() *regexp.Regexp
	ServiceProhibitedNameRegex() *regexp.Regexp

	RepositoryPermittedNameRegex() *regexp.Regexp
	RepositoryProhibitedNameRegex() *regexp.Regexp
	RepositoryTypes() []string
	RepositoryKeySeparator() string
}

// Custom is a type casting helper that gets you from the configuration acorn to your CustomConfiguration
func Custom(configuration librepo.Configuration) CustomConfiguration {
	return configuration.Custom().(CustomConfiguration)
}
