package sshAuthProvider

import (
	"context"
	"github.com/Interhyp/metadata-service/test/mock/configmock"
	"testing"

	"github.com/Interhyp/go-backend-service-common/docs"
	"github.com/stretchr/testify/require"
)

func TestProvideSshAuth(t *testing.T) {
	docs.Description("SshAuthProviderImpl works")

	sshAuthProvider := SshAuthProviderImpl{
		CustomConfiguration: new(configmock.MockConfig),
	}

	require.NotNil(t, sshAuthProvider)
	require.Equal(t, true, sshAuthProvider.IsSshAuthProvider())

	sshAuth, err := sshAuthProvider.ProvideSshAuth(context.Background())
	require.Nil(t, err)
	require.NotNil(t, sshAuth)
}
