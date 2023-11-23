package configmock

import (
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Roshick/go-autumn-kafka/pkg/aukafka"
	"regexp"
)

type MockConfig struct {
}

func (c *MockConfig) BasicAuthUsername() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) BasicAuthPassword() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) SSHPrivateKey() string {
	return `
-----BEGIN OPENSSH PRIVATE KEY-----
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
-----END OPENSSH PRIVATE KEY-----`
}

func (c *MockConfig) SSHPrivateKeyPassword() string {
	return "techexgeheim"
}

func (c *MockConfig) SSHMetadataRepositoryUrl() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) MetadataRepoUrl() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) BitbucketUsername() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) BitbucketPassword() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) BitbucketServer() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) BitbucketCacheSize() int {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) BitbucketCacheRetentionSeconds() uint32 {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) BitbucketReviewerFallback() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) GitCommitterName() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) GitCommitterEmail() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) KafkaUsername() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) KafkaPassword() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) KafkaTopic() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) KafkaSeedBrokers() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) KafkaGroupIdOverride() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) AuthOidcKeySetUrl() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) AuthOidcTokenAudience() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) AuthGroupWrite() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) UpdateJobIntervalCronPart() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) UpdateJobTimeoutSeconds() uint16 {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) AlertTargetPrefix() string {
	return "https://some-domain.com/"
}

func (c *MockConfig) AlertTargetSuffix() string {
	return "@some-organisation.com"
}

func (c *MockConfig) ElasticApmEnabled() bool {
	return false
}

func (c *MockConfig) OwnerAliasPermittedRegex() *regexp.Regexp {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) OwnerAliasProhibitedRegex() *regexp.Regexp {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) OwnerAliasMaxLength() uint16 {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) OwnerFilterAliasRegex() *regexp.Regexp {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) ServiceNamePermittedRegex() *regexp.Regexp {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) ServiceNameProhibitedRegex() *regexp.Regexp {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) ServiceNameMaxLength() uint16 {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) RepositoryNamePermittedRegex() *regexp.Regexp {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) RepositoryNameProhibitedRegex() *regexp.Regexp {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) RepositoryNameMaxLength() uint16 {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) RepositoryTypes() []string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) RepositoryKeySeparator() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) MetadataRepoMainline() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) NotificationConsumerConfigs() map[string]config.NotificationConsumerConfig {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) AllowedFileCategories() []string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) Kafka() *aukafka.Config {
	//TODO implement me
	panic("implement me")
}
