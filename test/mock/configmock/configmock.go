package configmock

import (
	"regexp"

	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Roshick/go-autumn-kafka/pkg/kafka"
)

type MockConfig struct {
}

func (c *MockConfig) UserPrefix() string {
	return ""
}

func (c *MockConfig) WebhooksProcessAsync() bool {
	//TODO implement me
	panic("implement me")
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

func (c *MockConfig) MetadataRepoUrl() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) ReviewerFallback() string {
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

func (c *MockConfig) AlertTargetRegex() *regexp.Regexp {
	return regexp.MustCompile("@some-organisation[.]com$")
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

func (c *MockConfig) Kafka() *kafka.Config {
	//TODO implement me
	panic("implement me")
}

func (c *MockConfig) RedisUrl() string {
	return ""
}

func (c *MockConfig) RedisPassword() string {
	return ""
}

func (c *MockConfig) MetadataRepoProject() string {
	return "sample"
}

func (c *MockConfig) MetadataRepoName() string {
	return "sample-repo"
}

func (c *MockConfig) PullRequestBuildUrl() string {
	return "https://example.com"
}

func (c *MockConfig) PullRequestBuildKey() string {
	return "metadata-service"
}

func (c *MockConfig) GithubAppId() int64 {
	return int64(1)
}

func (c *MockConfig) GithubAppInstallationId() int64 {
	return int64(1)
}

func (c *MockConfig) GithubAppJwtSigningKeyPEM() []byte {
	return []byte(`-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDWyiEOZQ1CEjRL
qysxSc4WMm7mNaQMndu9R45ZcmsimNAnH14J2Ooj2j5/andNBo51QiuRiJea2nZZ
/SLD4pcd4lxRbDvY7QhLY0O8MnpHg3V2DnsJctkR8LOwwuHRORyjCYMripltk9Cj
DeTwfU1AFuf9F2zYYbay03rWOc1exZFHC0eWEhJN9r0MVE99N0MVfGbb8l5BgfPP
BQH7/B1A8AlqqaVnPwGUBa2jw78e5edsLbQAPt/3FWKbkOshE52WbkCes021bUwj
5j8wJhi4+UmrUvNvELLi4+thp1tU/xZ+Lu880xm7ajF1DKXo/CHPEQ7HDrjfwcdk
2LdmgfJTAgMBAAECggEAXQD57ks4Qe8zAL7VvYpZN8hPt9PrPGFQKDXnP/joxfrI
SuBsrjPkMnEKVc6qaMpZfhGQXvx3tOA6lf2jg5FGYPTGh6UnhucgC9CoIEH1K6kS
//MGOJGnx3pjvDquYBNsQHZae0yQ4d863JekFbQT8pfYjQELKuionOcwjblKoWl8
YgiA496qVG18EOVnS3kHj5H1wJD2Xf3ptLKI+bjXAfXaiBn4fGdlqE4fHuZLHd8d
5lAcl5TU2s6G2KyXJyvMeD82/fUep+oTnRTHMtEqqDlFXmqKC6AIJm16t/IaGo4c
Ym87dbYJwHD+0kERMpMqykre/AlmWlL2Lq0lL8WtgQKBgQDcBqK8gR3tVgChRve7
cep5ocJYjm2RRBqbwzeOpM4tSnlJnlpIfGFLw3YFFGFsKja6aV7pr4LHk1EIslVo
y2lbQnRIEGk0jGx9PgSp4dd5lsAnW/wBnwmEBNhEN1nL3lya2lXfKwUTyEXNyaXX
vcXaiMt3fwzD/27SjvdoYMhogQKBgQD56FJTHqofl1K2I4n/nCMtGqxq2MzU1Gif
h4NVxpD2Gn70P3h0MX+0M9wfgT1T7JFMsI1VRazncsLoDDsb9r5+EPOYY0+wv4Uy
83awKUazglYGEBDHHRdbDJkx3gsp583aY73yJrGGh5IcuW0UfhY32mKukgcj5uSn
Wvn13uvQ0wKBgAbTSd8RHk2Lem+GVQ8ChKSLSQ0YNfvooe6tCp8pK6AqDEMlX2Wa
PiZshM+5hyAk2xfDRwd2w1bPkhbz+URL8xO6pwLJR4oyxPbJorlmYRnLfGB8MQAX
3+Kxh8ft86IoXrULCtjma7zmXIv6smNT5rxVvAIT9eBqnxR3DOO3BOCBAoGAKHNi
X/Hmt5ZW3QSDocw0JWjb36+X+BsplCjrKUcqz6saQY7EgIpCkXiTeMYCl0MDgdZS
CittAUmiIs1YA/68dstnopLwoztc5BJkc786onPGWNTg4lnjHem8IkY+qFnNCDx8
0mVQ9uWa0OtyrI58Ki4/KuKYJUeKW0xuiU27/eECgYBZS8SpocgTeHSs6tC4mYr/
GHC84dc4JrBll9zVtW3amw5+eUU31h48mEEFM4Sph4YlMIEenNiy0+6QAr3P212B
+r5dw0/D3o4wp7VYaieS11g2ZrMgLVFbKCvyH4rNdPn6QgSsxK22SnoPDkiJAbMS
0TEd3w/5KBsZU2kLdnQ0/Q==
-----END PRIVATE KEY-----`)
}

func (c *MockConfig) YamlIndentation() int {
	return 4
}

func (c *MockConfig) GithubAppWebhookSecret() []byte {
	return nil
}

func (c *MockConfig) FormattingActionCommitMsgPrefix() string {
	return ""
}

func (c *MockConfig) CheckWarnMissingMainlineProtection() bool {
	return false
}

func (c *MockConfig) CheckExpectedRequiredConditions() []config.CheckedRequiredConditions {
	return make([]config.CheckedRequiredConditions, 0)
}

func (c *MockConfig) CheckedExpectedExemptions() []config.CheckedExpectedExemption {
	return make([]config.CheckedExpectedExemption, 0)
}
