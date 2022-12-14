package idp

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/config"
	aurestclientprometheus "github.com/StephanHCB/go-autumn-restclient-prometheus"
	aurestclientapi "github.com/StephanHCB/go-autumn-restclient/api"
	auresthttpclient "github.com/StephanHCB/go-autumn-restclient/implementation/httpclient"
	aurestlogging "github.com/StephanHCB/go-autumn-restclient/implementation/requestlogging"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"net/http"
	"time"
)
import _ "github.com/go-git/go-git/v5"

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging

	IDPClient aurestclientapi.Client
	PEMKeySet []string
}

func (r *Impl) Setup(ctx context.Context) error {
	r.Logging.Logger().Ctx(ctx).Info().Print("setting up idp connector")

	client, err := auresthttpclient.New(10*time.Second, nil, nil)
	if err != nil {
		return err
	}
	aurestclientprometheus.InstrumentHttpClient(client)

	logWrapper := aurestlogging.New(client)

	r.IDPClient = logWrapper
	r.PEMKeySet = make([]string, 0)
	return nil
}

func (r *Impl) ObtainKeySet(ctx context.Context) error {
	keysetUrl := r.CustomConfiguration.AuthOidcKeySetUrl()

	responseMap := make(map[string]interface{})
	response := &aurestclientapi.ParsedResponse{
		Body: &responseMap,
	}

	err := r.IDPClient.Perform(ctx, http.MethodGet, keysetUrl, nil, response)
	if err != nil {
		return err
	}

	if response.Status != http.StatusOK {
		return errors.New("did not receive http 200 from idp")
	}

	// we have ensured a structured response, so it can't try to misinterpret e.g. blank pages, httpd error messages, ...
	keySetBytes, err := json.Marshal(&responseMap)
	if err != nil {
		return err
	}

	keySet, err := jwk.Parse(keySetBytes)
	if err != nil {
		return fmt.Errorf("failed to parse keyset: %v", err)
	}

	for i := 0; i < keySet.Len(); i++ {
		key, ok := keySet.Key(i)
		if !ok {
			return fmt.Errorf("failed to get key #%d from keyset", i+1)
		}

		pubKey := &rsa.PublicKey{}
		err = key.Raw(pubKey)
		if err != nil {
			return fmt.Errorf("failed to extract raw rsa public key for key #%d: %s", i+1, err.Error())
		}

		pubData, err := x509.MarshalPKIXPublicKey(pubKey)
		if err != nil {
			return fmt.Errorf("failed to marshal key #%d to public key: %s", i+1, err.Error())
		}

		output := bytes.Buffer{}
		err = pem.Encode(&output, &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubData,
		})
		if err != nil {
			return fmt.Errorf("failed to pem encode key #%d: %s", i+1, err.Error())
		}

		r.PEMKeySet = append(r.PEMKeySet, output.String())
	}

	return nil
}

func (r *Impl) GetKeySet(ctx context.Context) []string {
	return r.PEMKeySet
}

func (r *Impl) VerifyToken(ctx context.Context, token string) error {
	// TODO implement
	return nil
}
