package authProvider

import (
	"context"
	"github.com/Interhyp/metadata-service/test/mock/configmock"
	"github.com/Interhyp/metadata-service/test/mock/githubmock"
	auloggingapi "github.com/StephanHCB/go-autumn-logging/api"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"testing"

	"github.com/Interhyp/go-backend-service-common/docs"
	"github.com/stretchr/testify/require"
)

type MockLogging struct {
}

func (m MockLogging) IsLogging() bool {
	//TODO implement me
	panic("implement me")
}

func (m MockLogging) Setup() {
	//TODO implement me
	panic("implement me")
}

func (m MockLogging) Logger() auloggingapi.LoggingImplementation {
	return MockLoggingImplementation{}
}

type MockLoggingImplementation struct {
}

func (m MockLoggingImplementation) Ctx(ctx context.Context) auloggingapi.ContextAwareLoggingImplementation {
	return MockContextAwareLoggingImplementation{}
}

func (m MockLoggingImplementation) NoCtx() auloggingapi.ContextAwareLoggingImplementation {
	//TODO implement me
	panic("implement me")
}

type MockContextAwareLoggingImplementation struct {
}

func (m MockContextAwareLoggingImplementation) Trace() auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

func (m MockContextAwareLoggingImplementation) Debug() auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

func (m MockContextAwareLoggingImplementation) Info() auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

func (m MockContextAwareLoggingImplementation) Warn() auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

func (m MockContextAwareLoggingImplementation) Error() auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

func (m MockContextAwareLoggingImplementation) Fatal() auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

func (m MockContextAwareLoggingImplementation) Panic() auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

type MockLeveledLoggingImplementation struct {
}

func (m MockLeveledLoggingImplementation) WithErr(err error) auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

func (m MockLeveledLoggingImplementation) With(key string, value string) auloggingapi.LeveledLoggingImplementation {
	return MockLeveledLoggingImplementation{}
}

func (m MockLeveledLoggingImplementation) Print(v ...interface{}) {
}

func (m MockLeveledLoggingImplementation) Printf(format string, v ...interface{}) {
	// do nothing
}

func TestProvideAuth(t *testing.T) {
	docs.Description("AuthProviderImpl works")

	authProvider := AuthProviderImpl{
		CustomConfiguration: new(configmock.MockConfig),
		Logging:             MockLogging{},
		Github:              new(githubmock.GitHubMock),
	}

	err := authProvider.Setup()
	require.Nil(t, err)

	require.NotNil(t, authProvider)
	require.Equal(t, true, authProvider.IsAuthProvider())

	auth := authProvider.ProvideAuth(context.Background())
	require.NotNil(t, auth)
	if basicAuth, ok := auth.(*http.BasicAuth); ok {
		dereferencedBasicAuth := *basicAuth
		require.IsType(t, http.BasicAuth{}, dereferencedBasicAuth)
	} else {
		t.Errorf("Object expected to be of type http.BasicAuth, but was %T", auth)
	}
}
