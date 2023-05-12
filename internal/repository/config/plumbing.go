package config

import (
	"regexp"
	"strconv"

	"github.com/Interhyp/metadata-service/acorns/config"
	auacornapi "github.com/StephanHCB/go-autumn-acorn-registry/api"
	auconfigapi "github.com/StephanHCB/go-autumn-config-api"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
	libconfig "github.com/StephanHCB/go-backend-service-common/repository/config"
	"github.com/StephanHCB/go-backend-service-common/repository/vault"
)

type CustomConfigImpl struct {
	VBasicAuthUsername              string
	VBasicAuthPassword              string
	VSSHPrivateKey                  string
	VSSHPrivateKeyPassword          string
	VSSHMetadataRepoUrl             string
	VBitbucketUsername              string
	VBitbucketPassword              string
	VBitbucketServer                string
	VBitbucketCacheSize             int
	VBitbucketCacheRetentionSeconds uint32
	VBitbucketReviewerFallback      string
	VGitCommitterName               string
	VGitCommitterEmail              string
	VKafkaUsername                  string
	VKafkaPassword                  string
	VKafkaTopic                     string
	VKafkaSeedBrokers               string
	VAuthOidcKeySetUrl              string
	VAuthOidcTokenAudience          string
	VAuthGroupWrite                 string
	VKafkaGroupIdOverride           string
	VMetadataRepoMainline           string
	VUpdateJobIntervalCronPart      string
	VUpdateJobTimeoutSeconds        uint16
	VAlertTargetPrefix              string
	VAlertTargetSuffix              string
	VAdditionalPromoters            string
	VAdditionalPromotersFromOwners  string
	VElasticApmDisabled             bool
	VOwnerAliasPermittedRegex       *regexp.Regexp
	VOwnerAliasProhibitedRegex      *regexp.Regexp
	VOwnerAliasMaxLength            uint16
	VOwnerAliasFilterRegex          *regexp.Regexp
	VServiceNamePermittedRegex      *regexp.Regexp
	VServiceNameProhibitedRegex     *regexp.Regexp
	VServiceNameMaxLength           uint16
	VRepositoryNamePermittedRegex   *regexp.Regexp
	VRepositoryNameProhibitedRegex  *regexp.Regexp
	VRepositoryNameMaxLength        uint16
	VRepositoryTypes                string
	VRepositoryKeySeparator         string
}

func New() auacornapi.Acorn {
	instance := &CustomConfigImpl{}
	configItems := make([]auconfigapi.ConfigItem, 0)
	configItems = append(configItems, CustomConfigItems...)
	configItems = append(configItems, vault.ConfigItems...)
	return libconfig.New(instance, configItems)
}

func (c *CustomConfigImpl) Obtain(getter func(key string) string) {
	c.VBasicAuthUsername = getter(config.KeyBasicAuthUsername)
	c.VBasicAuthPassword = getter(config.KeyBasicAuthPassword)
	c.VSSHPrivateKey = getter(config.KeySSHPrivateKey)
	c.VSSHPrivateKeyPassword = getter(config.KeySSHPrivateKeyPassword)
	c.VSSHMetadataRepoUrl = getter(config.KeySSHMetadataRepositoryUrl)
	c.VBitbucketUsername = getter(config.KeyBitbucketUsername)
	c.VBitbucketPassword = getter(config.KeyBitbucketPassword)
	c.VBitbucketServer = getter(config.KeyBitbucketServer)
	c.VBitbucketCacheSize = toInt(getter(config.KeyBitbucketCacheSize))
	c.VBitbucketCacheRetentionSeconds = toUint32(getter(config.KeyBitbucketCacheRetentionSeconds))
	c.VBitbucketReviewerFallback = getter(config.KeyBitbucketReviewerFallback)
	c.VGitCommitterName = getter(config.KeyGitCommitterName)
	c.VGitCommitterEmail = getter(config.KeyGitCommitterEmail)
	c.VKafkaUsername = getter(config.KeyKafkaUsername)
	c.VKafkaPassword = getter(config.KeyKafkaPassword)
	c.VKafkaTopic = getter(config.KeyKafkaTopic)
	c.VKafkaSeedBrokers = getter(config.KeyKafkaSeedBrokers)
	c.VKafkaGroupIdOverride = getter(config.KeyKafkaGroupIdOverride)
	c.VAuthOidcKeySetUrl = getter(config.KeyAuthOidcKeySetUrl)
	c.VAuthOidcTokenAudience = getter(config.KeyAuthOidcTokenAudience)
	c.VAuthGroupWrite = getter(config.KeyAuthGroupWrite)
	c.VMetadataRepoMainline = getter(config.KeyMetadataRepoMainline)
	c.VUpdateJobIntervalCronPart = getter(config.KeyUpdateJobIntervalMinutes)
	c.VUpdateJobTimeoutSeconds = toUint16(getter(config.KeyUpdateJobTimeoutSeconds))
	c.VAlertTargetPrefix = getter(config.KeyAlertTargetPrefix)
	c.VAlertTargetSuffix = getter(config.KeyAlertTargetSuffix)
	c.VAdditionalPromoters = getter(config.KeyAdditionalPromoters)
	c.VAdditionalPromotersFromOwners = getter(config.KeyAdditionalPromotersFromOwners)
	c.VElasticApmDisabled, _ = strconv.ParseBool(getter(config.KeyElasticApmDisabled))
	c.VOwnerAliasPermittedRegex, _ = regexp.Compile(getter(config.KeyOwnerAliasPermittedRegex))
	c.VOwnerAliasProhibitedRegex, _ = regexp.Compile(getter(config.KeyOwnerAliasProhibitedRegex))
	c.VOwnerAliasMaxLength = toUint16(getter(config.KeyOwnerAliasMaxLength))
	c.VOwnerAliasFilterRegex, _ = regexp.Compile(getter(config.KeyOwnerAliasFilterRegex))
	c.VServiceNamePermittedRegex, _ = regexp.Compile(getter(config.KeyServiceNamePermittedRegex))
	c.VServiceNameProhibitedRegex, _ = regexp.Compile(getter(config.KeyServiceNameProhibitedRegex))
	c.VServiceNameMaxLength = toUint16(getter(config.KeyServiceNameMaxLength))
	c.VRepositoryNamePermittedRegex, _ = regexp.Compile(getter(config.KeyRepositoryNamePermittedRegex))
	c.VRepositoryNameProhibitedRegex, _ = regexp.Compile(getter(config.KeyRepositoryNameProhibitedRegex))
	c.VRepositoryNameMaxLength = toUint16(getter(config.KeyRepositoryNameMaxLength))
	c.VRepositoryTypes = getter(config.KeyRepositoryTypes)
	c.VRepositoryKeySeparator = getter(config.KeyRepositoryKeySeparator)
}

// used after validation, so known safe

func toInt(s string) int {
	val, _ := auconfigenv.AToInt(s)
	return val
}

func toUint16(s string) uint16 {
	val, _ := auconfigenv.AToUint(s)
	return uint16(val)
}

func toUint32(s string) uint32 {
	val, _ := auconfigenv.AToUint(s)
	return uint32(val)
}
