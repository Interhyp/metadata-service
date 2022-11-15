package vault

import (
	"context"
	"errors"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/config"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
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

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging

	VaultEnabled            bool
	VaultProtocol           string
	VaultServer             string
	VaultToken              string
	VaultKubernetesRole     string
	VaultKubernetesAuthPath string
	VaultKubernetesBackend  string
	VaultServiceSecretsPath string
	VaultKafkaSecretsPath   string

	VaultClient aurestclientapi.Client
}

func (v *Impl) Setup(ctx context.Context) error {
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

func (v *Impl) publicCertOrNil() ([]byte, error) {
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

func (v *Impl) vaultRequestHeaderManipulator() func(ctx context.Context, r *http.Request) {
	return func(ctx context.Context, r *http.Request) {
		r.Header.Set(headers.Accept, aurestclientapi.ContentTypeApplicationJson)
		if v.VaultToken != "" {
			r.Header.Set("X-Vault-Token", v.VaultToken)
		}
	}
}

type K8sAuthRequest struct {
	Jwt  string `json:"jwt"`
	Role string `json:"role"`
}

type K8sAuthResponse struct {
	Auth   *K8sAuth `json:"auth"`
	Errors []string `json:"errors"`
}

type K8sAuth struct {
	ClientToken string `json:"client_token"`
}

func (v *Impl) Authenticate(ctx context.Context) error {
	if v.VaultToken != "" {
		v.Logging.Logger().Ctx(ctx).Info().Print("using passed in vault token, skipping authentication with vault")
		return nil
	} else {
		v.Logging.Logger().Ctx(ctx).Info().Print("authenticating with vault")

		remoteUrl := fmt.Sprintf("%s://%s/v1/auth/%s/login", v.VaultProtocol, v.VaultServer, v.VaultKubernetesBackend)

		k8sToken, err := os.ReadFile(v.VaultKubernetesAuthPath)
		if err != nil {
			return fmt.Errorf("unable to read vault token file from path %s: %s", v.VaultKubernetesAuthPath, err.Error())
		}

		requestDto := &K8sAuthRequest{
			Jwt:  string(k8sToken),
			Role: v.VaultKubernetesRole,
		}

		responseDto := &K8sAuthResponse{}
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

		v.VaultToken = responseDto.Auth.ClientToken

		return nil
	}
}

type SecretsResponse struct {
	Data   *SecretsResponseData `json:"data"`
	Errors []string             `json:"errors"`
}

type SecretsResponseData struct {
	Data map[string]string `json:"data"`
}

func (v *Impl) ObtainSecrets(ctx context.Context) error {
	secrets, err := v.lowlevelObtainSecrets(ctx, v.VaultServiceSecretsPath)
	if err != nil {
		return err
	}

	auconfigenv.Set(config.KeyBasicAuthUsername, secrets["BASIC_AUTH_USERNAME"])
	auconfigenv.Set(config.KeyBasicAuthPassword, secrets["BASIC_AUTH_PASSWORD"])
	auconfigenv.Set(config.KeyBitbucketUsername, secrets["BITBUCKET_USERNAME"])
	auconfigenv.Set(config.KeyBitbucketPassword, secrets["BITBUCKET_PASSWORD"])

	return nil
}

func (v *Impl) ObtainKafkaSecrets(ctx context.Context) error {
	fullSecretsPath := v.VaultKafkaSecretsPath
	if fullSecretsPath == "" {
		v.Logging.Logger().Ctx(ctx).Info().Printf("NOT querying vault for kafka secret, configuration missing (ok, feature toggle)")
		return nil
	}

	secrets, err := v.lowlevelObtainSecrets(ctx, fullSecretsPath)
	if err != nil {
		return err
	}

	auconfigenv.Set(config.KeyKafkaUsername, secrets["KAFKA_USERNAME"])
	auconfigenv.Set(config.KeyKafkaPassword, secrets["KAFKA_PASSWORD"])

	return nil
}

func (v *Impl) lowlevelObtainSecrets(ctx context.Context, fullSecretsPath string) (map[string]string, error) {
	emptyMap := make(map[string]string)

	v.Logging.Logger().Ctx(ctx).Info().Printf("querying vault for secrets, secret path %s", fullSecretsPath)

	remoteUrl := fmt.Sprintf("%s://%s/v1/system_kv/data/v1/%s", v.VaultProtocol, v.VaultServer, fullSecretsPath)

	responseDto := &SecretsResponse{}
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
