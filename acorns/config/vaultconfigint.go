package config

type VaultConfiguration interface {
}

const (
	KeyVaultEnabled             = "VAULT_ENABLED"
	KeyVaultServer              = "VAULT_SERVER"
	KeyVaultToken               = "LOCAL_VAULT_TOKEN"
	KeyVaultKubernetesRole      = "VAULT_KUBERNETES_ROLE"
	KeyVaultKubernetesTokenPath = "VAULT_KUBERNETES_TOKEN_PATH"
	KeyVaultKubernetesBackend   = "VAULT_KUBERNETES_BACKEND"
	KeyVaultServiceSecretsPath  = "VAULT_SERVICE_SECRETS_PATH"
	KeyVaultKafkaSecretsPath    = "VAULT_KAFKA_SECRETS_PATH"
)
