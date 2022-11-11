package config

import (
	auacornapi "github.com/StephanHCB/go-autumn-acorn-registry/api"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
	libconfig "github.com/StephanHCB/go-backend-service-common/repository/config"
	"regexp"
	"strconv"
)

type CustomConfigImpl struct {
	VBbUser                        string
	VGitCommitterName              string
	VGitCommitterEmail             string
	VKafkaUser                     string
	VKafkaTopic                    string
	VKafkaSeedBrokers              string
	VKeySetUrl                     string
	VKafkaGroupIdOverride          string
	VMetadataRepoUrl               string
	VUpdateJobIntervalCronPart     string
	VUpdateJobTimeoutSeconds       uint16
	VVaultSecretsBasePath          string
	VVaultKafkaSecretPath          string
	VAlertTargetPrefix             string
	VAlertTargetSuffix             string
	VAdditionalPromoters           string
	VAdditionalPromotersFromOwners string
	VElasticApmDisabled            bool
	VOwnerAliasPermittedRegex      *regexp.Regexp
	VOwnerAliasProhibitedRegex     *regexp.Regexp
	VOwnerAliasMaxLength           uint16
	VOwnerAliasFilterRegex         *regexp.Regexp
	VServiceNamePermittedRegex     *regexp.Regexp
	VServiceNameProhibitedRegex    *regexp.Regexp
	VServiceNameMaxLength          uint16
	VRepositoryNamePermittedRegex  *regexp.Regexp
	VRepositoryNameProhibitedRegex *regexp.Regexp
	VRepositoryNameMaxLength       uint16
	VRepositoryTypes               string
	VRepositoryKeySeparator        string
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
	c.VUpdateJobIntervalCronPart = getter(KeyUpdateJobIntervalMinutes)
	c.VUpdateJobTimeoutSeconds = toUint16(getter(KeyUpdateJobTimeoutSeconds))
	c.VVaultSecretsBasePath = getter(KeyVaultSecretsBasePath)
	c.VVaultKafkaSecretPath = getter(KeyVaultKafkaSecretPath)
	c.VAlertTargetPrefix = getter(KeyAlertTargetPrefix)
	c.VAlertTargetSuffix = getter(KeyAlertTargetSuffix)
	c.VAdditionalPromoters = getter(KeyAdditionalPromoters)
	c.VAdditionalPromotersFromOwners = getter(KeyAdditionalPromotersFromOwners)
	c.VElasticApmDisabled, _ = strconv.ParseBool(getter(KeyElasticApmDisabled))
	c.VOwnerAliasPermittedRegex, _ = regexp.Compile(getter(KeyOwnerAliasPermittedRegex))
	c.VOwnerAliasProhibitedRegex, _ = regexp.Compile(getter(KeyOwnerAliasProhibitedRegex))
	c.VOwnerAliasMaxLength = toUint16(getter(KeyOwnerAliasMaxLength))
	c.VOwnerAliasFilterRegex, _ = regexp.Compile(getter(KeyOwnerAliasFilterRegex))
	c.VServiceNamePermittedRegex, _ = regexp.Compile(getter(KeyServiceNamePermittedRegex))
	c.VServiceNameProhibitedRegex, _ = regexp.Compile(getter(KeyServiceNameProhibitedRegex))
	c.VServiceNameMaxLength = toUint16(getter(KeyServiceNameMaxLength))
	c.VRepositoryNamePermittedRegex, _ = regexp.Compile(getter(KeyRepositoryNamePermittedRegex))
	c.VRepositoryNameProhibitedRegex, _ = regexp.Compile(getter(KeyRepositoryNameProhibitedRegex))
	c.VRepositoryNameMaxLength = toUint16(getter(KeyRepositoryNameMaxLength))
	c.VRepositoryTypes = getter(KeyRepositoryTypes)
	c.VRepositoryKeySeparator = getter(KeyRepositoryKeySeparator)
}

// used after validation, so known safe

func toUint16(s string) uint16 {
	val, _ := auconfigenv.AToUint(s)
	return uint16(val)
}
