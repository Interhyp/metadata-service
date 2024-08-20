package bitbucket

import (
	"context"
	libconfig "github.com/Interhyp/go-backend-service-common/repository/config"
	"github.com/Interhyp/go-backend-service-common/repository/logging"
	configint "github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/errors/httperror"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/repository/config"
	"github.com/Interhyp/metadata-service/test/mock/bbclientmock"
	"github.com/Interhyp/metadata-service/test/mock/vaultmock"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
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

func TestNewAndSetup(t *testing.T) {
	vault := &vaultmock.VaultImpl{}
	logger := &logging.LoggingImpl{}
	conf := tstConfig(t)
	customConf := configint.Custom(conf)
	cut := New(conf, customConf, logger, vault)

	lowLevel := &bbclientmock.BitbucketClientMock{}
	cut.(*Impl).LowLevel = lowLevel

	require.True(t, cut.IsBitbucket())

	err := cut.Setup()
	require.Nil(t, err)
}

const validConfigurationPath = "../resources/valid-config.yaml"

func tstConfig(t *testing.T) *libconfig.ConfigImpl {
	impl, _ := config.New()
	configImpl := impl.(*libconfig.ConfigImpl)
	auconfigenv.LocalConfigFileName = validConfigurationPath
	err := configImpl.Read()
	require.Nil(t, err)
	return configImpl
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
