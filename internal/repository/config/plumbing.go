package config

import (
	auacornapi "github.com/StephanHCB/go-autumn-acorn-registry/api"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
	libconfig "github.com/StephanHCB/go-backend-service-common/repository/config"
)

type CustomConfigImpl struct {
	VBbUser                    string
	VGitCommitterName          string
	VGitCommitterEmail         string
	VKafkaUser                 string
	VKafkaTopic                string
	VKafkaSeedBrokers          string
	VKeySetUrl                 string
	VKafkaGroupIdOverride      string
	VMetadataRepoUrl           string
	VOwnerRegex                string
	VUpdateJobIntervalCronPart string
	VUpdateJobTimeoutSeconds   uint16
	VVaultSecretsBasePath      string
	VVaultKafkaSecretPath      string
	VAlertTargetPrefix         string
	VAlertTargetSuffix         string
	VAdditionalPromoters       string
}

func New() auacornapi.Acorn {
	instance := &CustomConfigImpl{}
	return libconfig.New(instance, CustomConfigItems)
}

func (c *CustomConfigImpl) Obtain(getter func(key string) string) {
	c.VBbUser = getter(KeyBbUser)
	c.VGitCommitterName = getter(KeyGitCommitterName)
	c.VGitCommitterEmail = getter(KeyGitCommitterEmail)
	c.VKafkaUser = getter(KeyKafkaUser)
	c.VKafkaTopic = getter(KeyKafkaTopic)
	c.VKafkaSeedBrokers = getter(KeyKafkaSeedBrokers)
	c.VKafkaGroupIdOverride = getter(KeyKafkaGroupIdOverride)
	c.VKeySetUrl = getter(KeyKeySetUrl)
	c.VMetadataRepoUrl = getter(KeyMetadataRepoUrl)
	c.VOwnerRegex = getter(KeyOwnerRegex)
	c.VUpdateJobIntervalCronPart = getter(KeyUpdateJobIntervalMinutes)
	c.VUpdateJobTimeoutSeconds = toUint16(getter(KeyUpdateJobTimeoutSeconds))
	c.VVaultSecretsBasePath = getter(KeyVaultSecretsBasePath)
	c.VVaultKafkaSecretPath = getter(KeyVaultKafkaSecretPath)
	c.VAlertTargetPrefix = getter(KeyAlertTargetPrefix)
	c.VAlertTargetSuffix = getter(KeyAlertTargetSuffix)
	c.VAdditionalPromoters = getter(KeyAdditionalPromoters)
}

// used after validation, so known safe

func toUint16(s string) uint16 {
	val, _ := auconfigenv.AToUint(s)
	return uint16(val)
}
