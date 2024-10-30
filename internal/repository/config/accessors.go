package config

import (
	"os"
	"regexp"
	"strings"

	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Roshick/go-autumn-kafka/pkg/kafka"
)

func (c *CustomConfigImpl) BasicAuthUsername() string {
	return c.VBasicAuthUsername
}

func (c *CustomConfigImpl) BasicAuthPassword() string {
	return c.VBasicAuthPassword
}

func (c *CustomConfigImpl) SSHPrivateKey() string {
	return c.VSSHPrivateKey
}

func (c *CustomConfigImpl) SSHPrivateKeyPassword() string {
	return c.VSSHPrivateKeyPassword
}

func (c *CustomConfigImpl) SSHMetadataRepositoryUrl() string {
	return c.VSSHMetadataRepoUrl
}

func (c *CustomConfigImpl) BitbucketUsername() string {
	return c.VBitbucketUsername
}

func (c *CustomConfigImpl) BitbucketPassword() string {
	return c.VBitbucketPassword
}

func (c *CustomConfigImpl) BitbucketServer() string {
	return c.VBitbucketServer
}

func (c *CustomConfigImpl) BitbucketCacheSize() int {
	return c.VBitbucketCacheSize
}

func (c *CustomConfigImpl) BitbucketCacheRetentionSeconds() uint32 {
	return c.VBitbucketCacheRetentionSeconds
}

func (c *CustomConfigImpl) BitbucketReviewerFallback() string {
	return c.VBitbucketReviewerFallback
}

func (c *CustomConfigImpl) GitCommitterName() string {
	return c.VGitCommitterName
}

func (c *CustomConfigImpl) GitCommitterEmail() string {
	return c.VGitCommitterEmail
}

func (c *CustomConfigImpl) AuthOidcKeySetUrl() string {
	return c.VAuthOidcKeySetUrl
}

func (c *CustomConfigImpl) AuthOidcTokenAudience() string {
	return c.VAuthOidcTokenAudience
}

func (c *CustomConfigImpl) AuthGroupWrite() string {
	return c.VAuthGroupWrite
}

func (c *CustomConfigImpl) KafkaGroupIdOverride() string {
	return c.VKafkaGroupIdOverride
}

func (c *CustomConfigImpl) MetadataRepoUrl() string {
	return c.VMetadataRepoUrl
}

func (c *CustomConfigImpl) MetadataRepoMainline() string {
	return c.VMetadataRepoMainline
}

func (c *CustomConfigImpl) UpdateJobIntervalCronPart() string {
	return c.VUpdateJobIntervalCronPart
}

func (c *CustomConfigImpl) UpdateJobTimeoutSeconds() uint16 {
	return c.VUpdateJobTimeoutSeconds
}

func (c *CustomConfigImpl) AlertTargetRegex() *regexp.Regexp {
	return c.VAlertTargetRegex
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

func (c *CustomConfigImpl) NotificationConsumerConfigs() map[string]config.NotificationConsumerConfig {
	return c.VNotificationConsumerConfigs
}

func (c *CustomConfigImpl) AllowedFileCategories() []string {
	return c.VAllowedFileCategories
}

func (c *CustomConfigImpl) Kafka() *kafka.Config {
	return c.VKafkaConfig
}

func (c *CustomConfigImpl) RedisUrl() string {
	return c.VRedisUrl
}

func (c *CustomConfigImpl) RedisPassword() string {
	return c.VRedisPassword
}

func (c *CustomConfigImpl) MetadataRepoProject() string {
	sshUrl := c.SSHMetadataRepositoryUrl()
	if sshUrl != "" {
		match := c.GitUrlMatcher.FindStringSubmatch(sshUrl)
		if len(match) == 4 {
			return match[2]
		}
	}
	httpUrl := c.MetadataRepoUrl()
	if httpUrl != "" {
		match := c.GitUrlMatcher.FindStringSubmatch(httpUrl)
		if len(match) == 4 {
			return match[2]
		}
	}
	return ""
}

func (c *CustomConfigImpl) MetadataRepoName() string {
	sshUrl := c.SSHMetadataRepositoryUrl()
	if sshUrl != "" {
		match := c.GitUrlMatcher.FindStringSubmatch(sshUrl)
		if len(match) == 4 {
			return match[3]
		}
	}
	httpUrl := c.MetadataRepoUrl()
	if httpUrl != "" {
		match := c.GitUrlMatcher.FindStringSubmatch(httpUrl)
		if len(match) == 4 {
			return match[3]
		}
	}
	return ""
}

func (c *CustomConfigImpl) PullRequestBuildUrl() string {
	return c.VPullRequestBuildUrl
}

func (c *CustomConfigImpl) PullRequestBuildKey() string {
	return c.VPullRequestBuildKey
}

func (c *CustomConfigImpl) VCSConfigs() map[string]config.VCSConfig {
	return c.VVCSConfig
}

func (c *CustomConfigImpl) WebhooksProcessAsync() bool {
	return c.VWebhooksProcessAsync
}

func (c *CustomConfigImpl) UserPrefix() string {
	return c.VUserPrefix
}
