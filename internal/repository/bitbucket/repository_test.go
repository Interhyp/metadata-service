package bitbucket

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/errors/httperror"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/test/acceptance/bbclientmock"
	"github.com/Interhyp/metadata-service/test/acceptance/vaultmock"
	"github.com/StephanHCB/go-backend-service-common/repository/logging"
	"github.com/stretchr/testify/require"
	"testing"
)

func tstSetup() Impl {
	lowLevel := bbclientmock.BitbucketClientMock{}
	vault := vaultmock.VaultImpl{}
	logger := logging.LoggingImpl{}
	return Impl{
		Vault:    &vault,
		LowLevel: &lowLevel,
		Logging:  &logger,
	}
}

func TestGetBitbucketUser_Success(t *testing.T) {
	cut := tstSetup()

	result, err := cut.GetBitbucketUser(context.Background(), "some-user")

	require.NotNil(t, result)
	require.Nil(t, err)
	require.Equal(t, bbclientmock.MockBitbucketUser(), result)
}

func TestGetBitbucketUser_Error(t *testing.T) {
	cut := tstSetup()

	result, err := cut.GetBitbucketUser(context.Background(), bbclientmock.NOT_EXISITNG_USER)

	require.NotNil(t, result)
	require.NotNil(t, err)
	require.Equal(t, repository.BitbucketUser{}, result)
}

func TestGetBitbucketUsers_Success(t *testing.T) {
	cut := tstSetup()

	result, err := cut.GetBitbucketUsers(context.Background(), []string{"some-user", "other-user"})

	require.NotNil(t, result)
	require.Nil(t, err)
	require.Equal(t, []repository.BitbucketUser{bbclientmock.MockBitbucketUser(), bbclientmock.MockBitbucketUser()}, result)
}

func TestGetBitbucketUsers_UserNotFound(t *testing.T) {
	cut := tstSetup()

	result, err := cut.GetBitbucketUsers(context.Background(), []string{bbclientmock.NOT_EXISITNG_USER})

	require.NotNil(t, result)
	require.Nil(t, err)
	require.Equal(t, []repository.BitbucketUser{}, result)
}

func TestGetBitbucketUsers_OtherHttpError(t *testing.T) {
	cut := tstSetup()

	result, err := cut.GetBitbucketUsers(context.Background(), []string{bbclientmock.HTTP_ERROR_USER})

	require.NotNil(t, result)
	require.NotNil(t, err)
	require.True(t, httperror.Is(err))
	require.Equal(t, []repository.BitbucketUser{}, result)
}

func TestGetBitbucketUsers_Error(t *testing.T) {
	cut := tstSetup()

	result, err := cut.GetBitbucketUsers(context.Background(), []string{bbclientmock.OTHER_ERROR_USER})

	require.NotNil(t, result)
	require.NotNil(t, err)
	require.False(t, httperror.Is(err))
	require.Equal(t, []repository.BitbucketUser{}, result)
}

func TestFilterExistingUsernames_Success(t *testing.T) {
	cut := tstSetup()

	result, err := cut.FilterExistingUsernames(context.Background(), []string{"some-user", "other-user", bbclientmock.NOT_EXISITNG_USER})

	require.NotNil(t, result)
	require.Nil(t, err)
	require.Equal(t, []string{"mock-user", "mock-user"}, result)
}

func TestFilterExistingUsernames_Error(t *testing.T) {
	cut := tstSetup()

	result, err := cut.FilterExistingUsernames(context.Background(), []string{"some-user", "other-user", bbclientmock.OTHER_ERROR_USER})

	require.NotNil(t, result)
	require.NotNil(t, err)
	require.Equal(t, []string{}, result)
}
