package vault

import (
	"context"
	"errors"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/repository"
	aurestclientprometheus "github.com/StephanHCB/go-autumn-restclient-prometheus"
	aurestclientapi "github.com/StephanHCB/go-autumn-restclient/api"
	auresthttpclient "github.com/StephanHCB/go-autumn-restclient/implementation/httpclient"
	aurestlogging "github.com/StephanHCB/go-autumn-restclient/implementation/requestlogging"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/go-http-utils/headers"
	"net/http"
	"os"
	"time"
)

type VaultImpl struct {
	Configuration       librepo.Configuration
	CustomConfiguration repository.CustomConfiguration
	Logging             librepo.Logging

	vaultToken    string
	VaultProtocol string

	bbPassword    string
	kafkaPassword string

	basicAuthUsername string
	basicAuthPassword string

	VaultClient aurestclientapi.Client
}

func (v *VaultImpl) Setup(ctx context.Context) error {
	v.Logging.Logger().Ctx(ctx).Info().Print("setting up vault")

	publicCertBytes, err := v.publicCertOrNil()
	if err != nil {
		return err
	}

	client, err := auresthttpclient.New(15*time.Second, publicCertBytes, v.vaultRequestHeaderManipulator())
	if err != nil {
		return err
	}
	aurestclientprometheus.InstrumentHttpClient(client)

	logWrapper := aurestlogging.New(client)

	v.VaultClient = logWrapper
	return nil
}

func (v *VaultImpl) publicCertOrNil() ([]byte, error) {
	publicCertFilename := v.Configuration.VaultCertificateFile()

	if publicCertFilename != "" {
		publicCertBytes, err := os.ReadFile(publicCertFilename)
		if err != nil {
			return nil, err
		}
		return publicCertBytes, nil
	} else {
		return nil, nil
	}
}

func (v *VaultImpl) vaultRequestHeaderManipulator() func(ctx context.Context, r *http.Request) {
	return func(ctx context.Context, r *http.Request) {
		r.Header.Set(headers.Accept, aurestclientapi.ContentTypeApplicationJson)
		if v.vaultToken != "" {
			r.Header.Set("X-Vault-Token", v.vaultToken)
		}
	}
}

type VaultK8sAuthRequest struct {
	Jwt  string `json:"jwt"`
	Role string `json:"role"`
}

type VaultK8sAuthResponse struct {
	Auth   *VaultK8sAuth `json:"auth"`
	Errors []string      `json:"errors"`
}

type VaultK8sAuth struct {
	ClientToken string `json:"client_token"`
}

func (v *VaultImpl) Authenticate(ctx context.Context) error {
	if v.Configuration.LocalVault() {
		v.Logging.Logger().Ctx(ctx).Info().Print("using passed in vault token, skipping authentication with vault")
		v.vaultToken = v.Configuration.LocalVaultToken()
		return nil
	} else {
		v.Logging.Logger().Ctx(ctx).Info().Print("authenticating with vault")

		vaultServer := v.Configuration.VaultServer()
		k8sBackend := v.Configuration.VaultKubernetesBackend()

		remoteUrl := fmt.Sprintf("%s://%s/v1/auth/%s/login", v.VaultProtocol, vaultServer, k8sBackend)

		k8sTokenPath := v.Configuration.VaultKubernetesTokenPath()
		k8sRole := v.Configuration.VaultKubernetesRole()

		k8sToken, err := os.ReadFile(k8sTokenPath)
		if err != nil {
			return fmt.Errorf("unable to read vault token file from path %s: %s", k8sTokenPath, err.Error())
		}

		requestDto := &VaultK8sAuthRequest{
			Jwt:  string(k8sToken),
			Role: k8sRole,
		}

		responseDto := &VaultK8sAuthResponse{}
		response := &aurestclientapi.ParsedResponse{
			Body: responseDto,
		}

		err = v.VaultClient.Perform(ctx, http.MethodPost, remoteUrl, requestDto, response)
		if err != nil {
			return err
		}

		if response.Status != http.StatusOK {
			return errors.New("did not receive http 200 from vault")
		}

		if len(responseDto.Errors) > 0 {
			v.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("failed to authenticate with vault: %v", responseDto.Errors)
			return errors.New("got an errors array from vault")
		}

		if responseDto.Auth == nil || responseDto.Auth.ClientToken == "" {
			return errors.New("response from vault did not include a client_token")
		}

		v.vaultToken = responseDto.Auth.ClientToken

		return nil
	}
}

type VaultSecretsResponse struct {
	Data   *VaultSecretsResponseData `json:"data"`
	Errors []string                  `json:"errors"`
}

type VaultSecretsResponseData struct {
	Data map[string]string `json:"data"`
}

func (v *VaultImpl) ObtainSecrets(ctx context.Context) error {
	fullSecretsPath := fmt.Sprintf("%s/%s/%s", v.Configuration.Custom().(repository.CustomConfiguration).VaultSecretsBasePath(), v.Configuration.Environment(), v.Configuration.VaultSecretPath())

	secrets, err := v.lowlevelObtainSecrets(ctx, fullSecretsPath)
	if err != nil {
		return err
	}

	v.bbPassword = secrets["BB_PASSWORD"]
	v.basicAuthUsername = secrets["BASIC_AUTH_USERNAME"]
	v.basicAuthPassword = secrets["BASIC_AUTH_PASSWORD"]

	return nil
}

func (v *VaultImpl) ObtainKafkaSecrets(ctx context.Context) error {
	fullSecretsPath := v.CustomConfiguration.VaultKafkaSecretPath()
	if fullSecretsPath == "" {
		v.Logging.Logger().Ctx(ctx).Info().Printf("NOT querying vault for kafka secret, configuration missing (ok, feature toggle)")
		return nil
	}

	secrets, err := v.lowlevelObtainSecrets(ctx, fullSecretsPath)
	if err != nil {
		return err
	}

	v.kafkaPassword = secrets["key"]

	return nil
}

func (v *VaultImpl) lowlevelObtainSecrets(ctx context.Context, fullSecretsPath string) (map[string]string, error) {
	emptyMap := make(map[string]string)

	v.Logging.Logger().Ctx(ctx).Info().Printf("querying vault for secrets, secret path %s", fullSecretsPath)

	remoteUrl := fmt.Sprintf("%s://%s/v1/system_kv/data/v1/%s", v.VaultProtocol, v.Configuration.VaultServer(), fullSecretsPath)

	responseDto := &VaultSecretsResponse{}
	response := &aurestclientapi.ParsedResponse{
		Body: responseDto,
	}

	err := v.VaultClient.Perform(ctx, http.MethodGet, remoteUrl, nil, response)
	if err != nil {
		return emptyMap, err
	}

	if response.Status != http.StatusOK {
		return emptyMap, errors.New("did not receive http 200 from vault")
	}

	if len(responseDto.Errors) > 0 {
		v.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("failed to obtain secrets from vault: %v", responseDto.Errors)
		return emptyMap, errors.New("got an errors array from vault")
	}

	if responseDto.Data == nil {
		return emptyMap, errors.New("got no top level data structure from vault")
	}
	if responseDto.Data.Data == nil {
		return emptyMap, errors.New("got no second level data structure from vault")
	}

	return responseDto.Data.Data, nil
}
