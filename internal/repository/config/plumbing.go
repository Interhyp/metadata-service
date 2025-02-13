package config

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	libconfig "github.com/Interhyp/go-backend-service-common/repository/config"
	"github.com/Interhyp/go-backend-service-common/repository/vault"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	openapi "github.com/Interhyp/metadata-service/internal/types"
	"github.com/Roshick/go-autumn-kafka/pkg/kafka"
	auconfigapi "github.com/StephanHCB/go-autumn-config-api"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
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
	VAuthOidcKeySetUrl              string
	VAuthOidcTokenAudience          string
	VAuthGroupWrite                 string
	VKafkaGroupIdOverride           string
	VMetadataRepoUrl                string
	VMetadataRepoMainline           string
	VUpdateJobIntervalCronPart      string
	VUpdateJobTimeoutSeconds        uint16
	VAlertTargetRegex               *regexp.Regexp
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
	VNotificationConsumerConfigs    map[string]config.NotificationConsumerConfig
	VRedisUrl                       string
	VRedisPassword                  string
	VPullRequestBuildUrl            string
	VPullRequestBuildKey            string
	VVCSConfig                      map[string]config.VCSConfig
	VWebhooksProcessAsync           bool

	VKafkaConfig  *kafka.Config
	GitUrlMatcher *regexp.Regexp
}

func New() (librepo.Configuration, config.CustomConfiguration) {
	instance := &CustomConfigImpl{
		VKafkaConfig:  kafka.NewConfig(),
		GitUrlMatcher: regexp.MustCompile(`(/|:)([^/]+)/([^/]+).git$`),
	}
	configItems := make([]auconfigapi.ConfigItem, 0)
	configItems = append(configItems, CustomConfigItems...)
	configItems = append(configItems, vault.ConfigItems...)
	configItems = append(configItems, instance.VKafkaConfig.ConfigItems()...)

	libInstance := libconfig.NewNoAcorn(instance, configItems)

	// perform conversion to CustomConfig here
	return libInstance, config.Custom(libInstance)
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
	c.VKafkaGroupIdOverride = getter(config.KeyKafkaGroupIdOverride)
	c.VAuthOidcKeySetUrl = getter(config.KeyAuthOidcKeySetUrl)
	c.VAuthOidcTokenAudience = getter(config.KeyAuthOidcTokenAudience)
	c.VAuthGroupWrite = getter(config.KeyAuthGroupWrite)
	c.VMetadataRepoUrl = getter(config.KeyMetadataRepoUrl)
	c.VMetadataRepoMainline = getter(config.KeyMetadataRepoMainline)
	c.VUpdateJobIntervalCronPart = getter(config.KeyUpdateJobIntervalMinutes)
	c.VUpdateJobTimeoutSeconds = toUint16(getter(config.KeyUpdateJobTimeoutSeconds))
	c.VAlertTargetRegex, _ = regexp.Compile(getter(config.KeyAlertTargetRegex))
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
	c.VNotificationConsumerConfigs, _ = parseNotificationConsumerConfigs(getter(config.KeyNotificationConsumerConfigs))
	c.VRedisUrl = getter(config.KeyRedisUrl)
	c.VRedisPassword = getter(config.KeyRedisPassword)
	c.VPullRequestBuildUrl = getter(config.KeyPullRequestBuildUrl)
	c.VPullRequestBuildKey = getter(config.KeyPullRequestBuildKey)
	c.VVCSConfig, _ = parseVCSConfigs(getter(config.KeyVCSConfigs), getter)
	c.VWebhooksProcessAsync, _ = toBoolean(getter(config.KeyWebhooksProcessAsync))

	c.VKafkaConfig.Obtain(getter)
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

func toBoolean(value string) (bool, error) {
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("value %s is not a valid boolean", value)
	}
	return boolValue, nil
}

func parseNotificationConsumerConfigs(rawJson string) (map[string]config.NotificationConsumerConfig, error) {
	result := make(map[string]config.NotificationConsumerConfig)
	if rawJson == "" {
		return result, nil
	}

	type StringBasedConfig struct {
		SubscribedTypes map[string][]string `json:"types"`
		ConsumerURL     *string             `json:"url"`
	}
	parsedConfig := make(map[string]StringBasedConfig)
	if err := json.Unmarshal([]byte(rawJson), &parsedConfig); err != nil {
		return nil, err
	}

	errors := make([]string, 0)
	for configIdentifier, stringConfig := range parsedConfig {
		if stringConfig.ConsumerURL == nil {
			errors = append(errors, fmt.Sprintf("Notification consumer config '%s' is missing url.", configIdentifier))
			continue
		}
		consumerUrl := *stringConfig.ConsumerURL
		if consumerUrl == "" || !strings.HasPrefix(consumerUrl, "http") {
			errors = append(errors, fmt.Sprintf("Notification consumer config '%s' contains invalid url '%s'.", configIdentifier, consumerUrl))
		}

		types := make(map[openapi.NotificationPayloadType]map[openapi.NotificationEventType]struct{})
		for typeCandidate, eventCandidates := range stringConfig.SubscribedTypes {
			var key openapi.NotificationPayloadType

			switch typeCandidate {
			case openapi.OwnerPayload.String():
				key = openapi.OwnerPayload
				break
			case openapi.ServicePayload.String():
				key = openapi.ServicePayload
				break
			case openapi.RepositoryPayload.String():
				key = openapi.RepositoryPayload
				break
			default:
				errors = append(errors, fmt.Sprintf("Notification consumer config '%s' contains invalid type '%s'.", configIdentifier, typeCandidate))
				continue
			}

			types[key] = make(map[openapi.NotificationEventType]struct{})
			for _, eventCandidate := range eventCandidates {
				switch eventCandidate {
				case openapi.CreatedEvent.String():
					types[key][openapi.CreatedEvent] = struct{}{}
					break
				case openapi.ModifiedEvent.String():
					types[key][openapi.ModifiedEvent] = struct{}{}
					break
				case openapi.DeletedEvent.String():
					types[key][openapi.DeletedEvent] = struct{}{}
					break
				default:
					errors = append(errors, fmt.Sprintf("Notification consumer config '%s' contains invalid event type '%s'.", configIdentifier, eventCandidate))
					continue
				}
			}
		}

		result[configIdentifier] = config.NotificationConsumerConfig{
			ConsumerURL: consumerUrl,
			Subscribed:  types,
		}
	}
	if len(errors) > 0 {
		return nil, fmt.Errorf(strings.Join(errors, " "))
	}
	return result, nil
}

type rawVCSConfig struct {
	Platform          string  `json:"platform"`
	APIBaseURL        string  `json:"apiBaseURL"`
	AccessToken       *string `json:"accessToken,omitempty"`
	AccessTokenEnvVar *string `json:"accessTokenEnvVar,omitempty"`
}

func parseVCSConfigs(jsonString string, getter func(string) string) (map[string]config.VCSConfig, error) {
	var vcsConfigsRaw map[string]rawVCSConfig
	if err := json.Unmarshal([]byte(jsonString), &vcsConfigsRaw); err != nil {
		return nil, err
	}

	vcsConfigs := make(map[string]config.VCSConfig, 0)
	for key, raw := range vcsConfigsRaw {
		var accessToken string
		// We do not support accessing vcs without access token
		if raw.AccessTokenEnvVar != nil {
			accessToken = getter(*raw.AccessTokenEnvVar)
			if accessToken == "" {
				return nil, fmt.Errorf("vcs %s: access-token variable %s is empty", key, *raw.AccessTokenEnvVar)
			}
		} else if raw.AccessToken != nil && *raw.AccessToken != "" {
			accessToken = *raw.AccessToken
		} else {
			return nil, fmt.Errorf("vcs %s: neither access-token environment variable or access-token value is set", key)
		}

		platform, err := parseVCSPlatform(raw.Platform)
		if err != nil {
			return nil, fmt.Errorf("vcs %s: %s", key, err.Error())
		}

		vcsConfigs[key] = config.VCSConfig{
			Platform:    platform,
			APIBaseURL:  raw.APIBaseURL,
			AccessToken: accessToken,
		}
	}
	return vcsConfigs, nil
}

func parseVCSPlatform(value string) (config.VCSPlatform, error) {
	switch value {
	case "BITBUCKET_DATACENTER":
		return config.VCSPlatformBitbucketDatacenter, nil
	case "GITHUB":
		return config.VCSPlatformGitHub, nil
	default:
		return config.VCSPlatformUnknown, fmt.Errorf("invalid vcs platform: '%s'", value)
	}
}
