package config

import (
	"os"
	"regexp"
	"strings"
)

func (c *CustomConfigImpl) BbUser() string {
	return c.VBbUser
}

func (c *CustomConfigImpl) GitCommitterName() string {
	return c.VGitCommitterName
}

func (c *CustomConfigImpl) GitCommitterEmail() string {
	return c.VGitCommitterEmail
}

func (c *CustomConfigImpl) KafkaUser() string {
	return c.VKafkaUser
}

func (c *CustomConfigImpl) KafkaTopic() string {
	return c.VKafkaTopic
}

func (c *CustomConfigImpl) KafkaSeedBrokers() string {
	return c.VKafkaSeedBrokers
}

func (c *CustomConfigImpl) KeySetUrl() string {
	return c.VKeySetUrl
}

func (c *CustomConfigImpl) KafkaGroupIdOverride() string {
	return c.VKafkaGroupIdOverride
}

func (c *CustomConfigImpl) MetadataRepoUrl() string {
	return c.VMetadataRepoUrl
}

func (c *CustomConfigImpl) UpdateJobIntervalCronPart() string {
	return c.VUpdateJobIntervalCronPart
}

func (c *CustomConfigImpl) UpdateJobTimeoutSeconds() uint16 {
	return c.VUpdateJobTimeoutSeconds
}

func (c *CustomConfigImpl) VaultSecretsBasePath() string {
	return c.VVaultSecretsBasePath
}

func (c *CustomConfigImpl) VaultKafkaSecretPath() string {
	return c.VVaultKafkaSecretPath
}

func (c *CustomConfigImpl) AlertTargetPrefix() string {
	return c.VAlertTargetPrefix
}

func (c *CustomConfigImpl) AlertTargetSuffix() string {
	return c.VAlertTargetSuffix
}

func (c *CustomConfigImpl) AdditionalPromoters() []string {
	return strings.Split(c.VAdditionalPromoters, ",")
}

func (c *CustomConfigImpl) AdditionalPromotersFromOwners() []string {
	return strings.Split(c.VAdditionalPromotersFromOwners, ",")
}

func (c *CustomConfigImpl) ElasticApmEnabled() bool {
	return !c.VElasticApmDisabled &&
		os.Getenv("ELASTIC_APM_SERVER_URL") != "" &&
		os.Getenv("ELASTIC_APM_SERVICE_NAME") != "" &&
		os.Getenv("ELASTIC_APM_ENVIRONMENT") != ""
}

func (c *CustomConfigImpl) OwnerAliasPermittedRegex() *regexp.Regexp {
	return c.VOwnerAliasPermittedRegex
}

func (c *CustomConfigImpl) OwnerAliasProhibitedRegex() *regexp.Regexp {
	return c.VOwnerAliasProhibitedRegex
}

func (c *CustomConfigImpl) OwnerAliasMaxLength() uint16 {
	return c.VOwnerAliasMaxLength
}

func (c *CustomConfigImpl) OwnerFilterAliasRegex() *regexp.Regexp {
	return c.VOwnerAliasFilterRegex
}

func (c *CustomConfigImpl) ServiceNamePermittedRegex() *regexp.Regexp {
	return c.VServiceNamePermittedRegex
}

func (c *CustomConfigImpl) ServiceNameProhibitedRegex() *regexp.Regexp {
	return c.VServiceNameProhibitedRegex
}

func (c *CustomConfigImpl) ServiceNameMaxLength() uint16 {
	return c.VServiceNameMaxLength
}

func (c *CustomConfigImpl) RepositoryNamePermittedRegex() *regexp.Regexp {
	return c.VRepositoryNamePermittedRegex
}

func (c *CustomConfigImpl) RepositoryNameProhibitedRegex() *regexp.Regexp {
	return c.VRepositoryNameProhibitedRegex
}

func (c *CustomConfigImpl) RepositoryNameMaxLength() uint16 {
	return c.VRepositoryNameMaxLength
}

func (c *CustomConfigImpl) RepositoryTypes() []string {
	return strings.Split(c.VRepositoryTypes, ",")
}

func (c *CustomConfigImpl) RepositoryKeySeparator() string {
	return c.VRepositoryKeySeparator
}
