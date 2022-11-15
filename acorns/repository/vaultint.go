package repository

import "context"

const VaultAcornName = "vault"

// Vault is the central singleton representing Hashicorp Vault.
//
// We use Vault to obtain sensitive configuration values, called "secrets".
type Vault interface {
	IsVault() bool

	// Setup uses the configuration
	Setup(ctx context.Context) error

	// Authenticate authenticates against vault
	Authenticate(ctx context.Context) error

	// ObtainSecrets fetches the regular secrets from vault
	ObtainSecrets(ctx context.Context) error

	// ObtainKafkaSecrets fetches the kafka secrets from vault (skipped if kafka username / topic not configured)
	ObtainKafkaSecrets(ctx context.Context) error
}
