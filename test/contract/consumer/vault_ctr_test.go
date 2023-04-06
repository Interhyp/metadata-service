package consumer

import (
	"context"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/config"
	auacorn "github.com/StephanHCB/go-autumn-acorn-registry"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/StephanHCB/go-backend-service-common/repository/vault"
	"github.com/StephanHCB/go-backend-service-common/web/util/media"
	"github.com/go-http-utils/headers"
	"github.com/pact-foundation/pact-go/dsl"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"testing"
)

func TestVaultConsumer_LocalToken_Success(t *testing.T) {
	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Consumer: "metadata",
		Provider: "vault",
		Host:     "localhost",
	}
	defer pact.Teardown()

	// The actual test case (consumer side)
	// This uses the repository on the consumer side to make the http call, should be as low level as possible
	var test = func() error {
		if err := tstSetup(); err != nil {
			return err
		}

		ctx := log.Logger.WithContext(context.Background())

		registry := auacorn.Registry.(*auacorn.AcornRegistryImpl)
		vaultImpl := registry.GetAcornByName(librepo.VaultAcornName).(*vault.Impl)

		if err := vaultImpl.Validate(ctx); err != nil {
			return err
		}
		vaultImpl.Obtain(ctx)
		// override target protocol and address
		vaultImpl.VaultServer = fmt.Sprintf("localhost:%d", pact.Server.Port)
		vaultImpl.VaultProtocol = "http"
		if err := vaultImpl.Setup(ctx); err != nil {
			return err
		}
		if err := vaultImpl.Authenticate(ctx); err != nil {
			return err
		}
		if err := vaultImpl.ObtainSecrets(ctx); err != nil {
			return err
		}

		require.Equal(t, "bb-secret-demosecret", auconfigenv.Get(config.KeyBitbucketPassword))

		return nil
	}

	// Set up our expected interactions.
	pact.
		AddInteraction().
		// contrived example, not really needed. This is the identifier of the state handler that will be called on the other side
		Given("an authorized user exists").
		UponReceiving("A request for the secrets").
		WithRequest(dsl.Request{
			Method: http.MethodGet,
			Headers: dsl.MapMatcher{
				headers.Accept:  dsl.String(media.ContentTypeApplicationJson),
				"X-Vault-Token": dsl.String("notarealtoken"),
			},
			Path: dsl.String("/v1/system_kv/data/v1/base/path/feat/some-service/secrets"),
		}).
		WillRespondWith(dsl.Response{
			Status:  200,
			Headers: dsl.MapMatcher{headers.ContentType: dsl.String(media.ContentTypeApplicationJson)},
			Body: map[string]interface{}{
				"request_id":     "2f724c34-406c-1e39-542d-670d662267fa",
				"lease_id":       "",
				"renewable":      false,
				"lease_duration": 0,
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"BB_PASSWORD": "bb-secret-demosecret",
					},
					"metadata": map[string]interface{}{
						"created_time":    "2021-08-13T06:43:45.831705283Z",
						"custom_metadata": nil,
						"deletion_time":   "",
						"destroyed":       false,
						"version":         2,
					},
				},
				"wrap_info": nil,
				"warnings":  nil,
				"auth":      nil,
			},
		})

	// Run the test, verify it did what we expected and capture the contract (writes a test log to logs/pact.log)
	if err := pact.Verify(test); err != nil {
		log.Error().Err(err).Msg("pact verify failed")
		logfile, _ := os.ReadFile("logs/pact.log")
		log.Error().Msg("pact log was (may wish to delete before running test):\n" + string(logfile))
		t.FailNow()
	}

	// now write out the contract json (by default it goes to subdirectory pacts)
	if err := pact.WritePact(); err != nil {
		log.Error().Err(err).Msg("write pact failed")
		t.FailNow()
	}
}
