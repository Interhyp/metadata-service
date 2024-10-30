package vcswebhookshandler

import "fmt"

type ErrVCSConfigurationNotFound struct {
	key string
}

func NewErrVCSConfigurationNotFound(key string) error {
	return ErrVCSConfigurationNotFound{
		key: key,
	}
}

func (e ErrVCSConfigurationNotFound) Error() string {
	return fmt.Sprintf("failed to find vcs configuration for key '%s'", e.key)
}
