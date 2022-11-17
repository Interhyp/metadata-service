package config

type VaultConfiguration interface {
}

const (
	KeyVaultEnabled                 = "VAULT_ENABLED"
	KeyVaultServer                  = "VAULT_SERVER"
	KeyVaultAuthToken               = "VAULT_AUTH_TOKEN"
	KeyVaultAuthKubernetesRole      = "VAULT_AUTH_KUBERNETES_ROLE"
	KeyVaultAuthKubernetesTokenPath = "VAULT_AUTH_KUBERNETES_TOKEN_PATH"
	KeyVaultAuthKubernetesBackend   = "VAULT_AUTH_KUBERNETES_BACKEND"
	KeyVaultSecretsConfig           = "VAULT_SECRETS_CONFIG"
)
