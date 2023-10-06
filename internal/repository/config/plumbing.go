package config

import (
	"encoding/json"
	"fmt"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	openapi "github.com/Interhyp/metadata-service/internal/types"
	auconfigapi "github.com/StephanHCB/go-autumn-config-api"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	libconfig "github.com/StephanHCB/go-backend-service-common/repository/config"
	"github.com/StephanHCB/go-backend-service-common/repository/vault"
	"regexp"
	"strconv"
	"strings"
)

const (
	allowedNotificationTypeService    = "Service"
	allowedNotificationTypeOwner      = "Owner"
	allowedNotificationTypeRepository = "Repository"
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
	VMetadataRepoUrl                string
	VMetadataRepoMainline           string
	VUpdateJobIntervalCronPart      string
	VUpdateJobTimeoutSeconds        uint16
	VAlertTargetPrefix              string
	VAlertTargetSuffix              string
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
}

func New() (librepo.Configuration, config.CustomConfiguration) {
	instance := &CustomConfigImpl{}
	configItems := make([]auconfigapi.ConfigItem, 0)
	configItems = append(configItems, CustomConfigItems...)
	configItems = append(configItems, vault.ConfigItems...)

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
	c.VKafkaUsername = getter(config.KeyKafkaUsername)
	c.VKafkaPassword = getter(config.KeyKafkaPassword)
	c.VKafkaTopic = getter(config.KeyKafkaTopic)
	c.VKafkaSeedBrokers = getter(config.KeyKafkaSeedBrokers)
	c.VKafkaGroupIdOverride = getter(config.KeyKafkaGroupIdOverride)
	c.VAuthOidcKeySetUrl = getter(config.KeyAuthOidcKeySetUrl)
	c.VAuthOidcTokenAudience = getter(config.KeyAuthOidcTokenAudience)
	c.VAuthGroupWrite = getter(config.KeyAuthGroupWrite)
	c.VMetadataRepoUrl = getter(config.KeyMetadataRepoUrl)
	c.VMetadataRepoMainline = getter(config.KeyMetadataRepoMainline)
	c.VUpdateJobIntervalCronPart = getter(config.KeyUpdateJobIntervalMinutes)
	c.VUpdateJobTimeoutSeconds = toUint16(getter(config.KeyUpdateJobTimeoutSeconds))
	c.VAlertTargetPrefix = getter(config.KeyAlertTargetPrefix)
	c.VAlertTargetSuffix = getter(config.KeyAlertTargetSuffix)
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
