package vault

// usable after Setup() - Authenticate() - ObtainSecrets()

func (v *VaultImpl) BbPassword() string {
	return v.bbPassword
}

func (v *VaultImpl) KafkaPassword() string {
	return v.kafkaPassword
}

func (v *VaultImpl) BasicAuthUsername() string {
	return v.basicAuthUsername
}

func (v *VaultImpl) BasicAuthPassword() string {
	return v.basicAuthPassword
}

// when adding a secret, you will need to
// - add a field to VaultImpl
// - add an accessor method to the repository.Vault interface and implement it here
// - set the field from the vault response at the end of ObtainSecrets()
