package config

import "strings"

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

func (c *CustomConfigImpl) OwnerRegex() string {
	return c.VOwnerRegex
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

func (c *CustomConfigImpl) AdditionalPromotersFromOwners() []string {
	return strings.Split(c.VAdditionalPromoters, ",")
}
