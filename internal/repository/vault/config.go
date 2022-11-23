package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/config"
	"github.com/Interhyp/metadata-service/acorns/repository"
	auconfigapi "github.com/StephanHCB/go-autumn-config-api"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
	"strconv"
)

var ConfigItems = []auconfigapi.ConfigItem{
	{
		Key:         config.KeyVaultEnabled,
		EnvName:     config.KeyVaultEnabled,
		Default:     "true",
		Description: "enables vault. supports all values supported by ParseBool (https://pkg.go.dev/strconv#ParseBool).",
		Validate:    auconfigenv.ObtainIsBooleanValidator(),
	},
	{
		Key:         config.KeyVaultAuthToken,
		EnvName:     config.KeyVaultAuthToken,
		Default:     "",
		Description: "authentication token used to fetch secrets.",
		Validate:    auconfigapi.ConfigNeedsNoValidation,
	},
	{
		Key:         config.KeyVaultAuthKubernetesRole,
		EnvName:     config.KeyVaultAuthKubernetesRole,
		Default:     "",
		Description: "role binding to use for vault kubernetes authentication.",
		Validate:    auconfigapi.ConfigNeedsNoValidation,
	},
	{
		Key:         config.KeyVaultAuthKubernetesTokenPath,
		EnvName:     config.KeyVaultAuthKubernetesTokenPath,
		Default:     "/var/run/secrets/kubernetes.io/serviceaccount/token",
		Description: "file path to the service-account token",
		Validate:    auconfigapi.ConfigNeedsNoValidation,
	},
	{
		Key:         config.KeyVaultAuthKubernetesBackend,
		EnvName:     config.KeyVaultAuthKubernetesBackend,
		Default:     "",
		Description: "authentication path for the kubernetes cluster",
		Validate:    auconfigapi.ConfigNeedsNoValidation,
	},
	{
		Key:         config.KeyVaultSecretsConfig,
		EnvName:     config.KeyVaultSecretsConfig,
		Default:     "{}",
		Description: "configuration consisting of vault paths and keys to fetch from the corresponding path. values will be written to the global configuration object.",
		Validate: func(key string) error {
			value := auconfigenv.Get(key)
			_, err := parseSecretsConfig(value)
			return err
		},
	},
}

func (v *Impl) Validate(ctx context.Context) error {
	var errorList = make([]error, 0)
	for _, it := range ConfigItems {
		if it.Validate != nil {
			err := it.Validate(it.Key)
			if err != nil {
				v.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("failed to validate configuration field %s", it.EnvName)
				errorList = append(errorList, err)
			}
		}
	}

	if len(errorList) > 0 {
		return fmt.Errorf("some configuration values failed to validate or parse. There were %d error(s). See details above", len(errorList))
	} else {
		return nil
	}
}

func (v *Impl) Obtain(ctx context.Context) {
	v.VaultEnabled, _ = strconv.ParseBool(auconfigenv.Get(config.KeyVaultEnabled))
	v.VaultServer = auconfigenv.Get(config.KeyVaultServer)
	v.VaultAuthToken = auconfigenv.Get(config.KeyVaultAuthToken)
	v.VaultAuthKubernetesRole = auconfigenv.Get(config.KeyVaultAuthKubernetesRole)
	v.VaultAuthKubernetesTokenPath = auconfigenv.Get(config.KeyVaultAuthKubernetesTokenPath)
	v.VaultAuthKubernetesBackend = auconfigenv.Get(config.KeyVaultAuthKubernetesBackend)
	v.VaultSecretsConfig, _ = parseSecretsConfig(auconfigenv.Get(config.KeyVaultSecretsConfig))
}

func parseSecretsConfig(jsonString string) (repository.VaultSecretsConfig, error) {
	secretsConfig := repository.VaultSecretsConfig{}
	if err := json.Unmarshal([]byte(jsonString), &secretsConfig); err != nil {
		return nil, err
	}
	return secretsConfig, nil
}
