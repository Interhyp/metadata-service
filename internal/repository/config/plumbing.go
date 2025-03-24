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
	VBasicAuthUsername             string
	VBasicAuthPassword             string
	VReviewerFallback              string
	VGitCommitterName              string
	VGitCommitterEmail             string
	VAuthOidcKeySetUrl             string
	VAuthOidcTokenAudience         string
	VAuthGroupWrite                string
	VKafkaGroupIdOverride          string
	VMetadataRepoUrl               string
	VMetadataRepoMainline          string
	VUpdateJobIntervalCronPart     string
	VUpdateJobTimeoutSeconds       uint16
	VAlertTargetRegex              *regexp.Regexp
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
	VNotificationConsumerConfigs   map[string]config.NotificationConsumerConfig
	VRedisUrl                      string
	VRedisPassword                 string
	VPullRequestBuildUrl           string
	VPullRequestBuildKey           string
	VWebhooksProcessAsync          bool
	VGithubAppId                   int64
	VGithubAppInstallationId       int64
	VGithubAppJwtSigningKeyPEM     []byte

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
	c.VReviewerFallback = getter(config.KeyReviewerFallback)
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
	c.VWebhooksProcessAsync, _ = toBoolean(getter(config.KeyWebhooksProcessAsync))
	c.VGithubAppId, _ = strconv.ParseInt(getter(config.KeyGithubAppId), 10, 64)
	c.VGithubAppInstallationId, _ = strconv.ParseInt(getter(config.KeyGithubAppInstallationId), 10, 64)
	c.VGithubAppJwtSigningKeyPEM = []byte(getter(config.KeyGithubAppJwtSigningKeyPEM))
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
