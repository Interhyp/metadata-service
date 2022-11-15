package vault

import (
	"context"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/config"
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
		Key:         config.KeyVaultServiceSecretsPath,
		EnvName:     config.KeyVaultServiceSecretsPath,
		Default:     "",
		Description: "total vault secret path is composed of VAULT_SECRETS_BASE_PATH/ENVIRONMENT/VAULT_SECRET_PATH",
		Validate:    auconfigenv.ObtainPatternValidator("^(|[a-z0-9-/]+)$"),
	},
	{
		Key:         config.KeyVaultKafkaSecretsPath,
		EnvName:     config.KeyVaultKafkaSecretsPath,
		Default:     "",
		Description: "optional: kafka secret path in vault (needed to send kafka notifications), leaving this or any of the other *KAFKA* fields empty will switch off all Kafka functionality, including the Vault query for Kafka credentials",
		Validate:    auconfigenv.ObtainPatternValidator("^(|[a-z0-9-/]+)$"),
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
	v.VaultToken = auconfigenv.Get(config.KeyVaultToken)
	v.VaultKubernetesRole = auconfigenv.Get(config.KeyVaultKubernetesRole)
	v.VaultKubernetesAuthPath = auconfigenv.Get(config.KeyVaultKubernetesTokenPath)
	v.VaultKubernetesBackend = auconfigenv.Get(config.KeyVaultKubernetesBackend)
	v.VaultServiceSecretsPath = auconfigenv.Get(config.KeyVaultServiceSecretsPath)
	v.VaultKafkaSecretsPath = auconfigenv.Get(config.KeyVaultKafkaSecretsPath)
}
