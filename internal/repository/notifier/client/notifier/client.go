package notifierclient

import (
	"context"
	"fmt"
	openapi "github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"net/http"
	"time"

	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	aurestbreakerprometheus "github.com/StephanHCB/go-autumn-restclient-circuitbreaker-prometheus"
	aurestbreaker "github.com/StephanHCB/go-autumn-restclient-circuitbreaker/implementation/breaker"
	aurestclientprometheus "github.com/StephanHCB/go-autumn-restclient-prometheus"
	aurestclientapi "github.com/StephanHCB/go-autumn-restclient/api"
	auresthttpclient "github.com/StephanHCB/go-autumn-restclient/implementation/httpclient"
	aurestlogging "github.com/StephanHCB/go-autumn-restclient/implementation/requestlogging"
)

type NotifierClient interface {
	Setup(clientIdentifier string, url string) error

	// Send will log any errors, but since we use it async, it cannot return the error
	Send(ctx context.Context, notification openapi.Notification)
}

type Impl struct {
	Logging             librepo.Logging
	CustomConfiguration config.CustomConfiguration

	clientIdentifier string
	url              string

	Client aurestclientapi.Client
}

func New(logging librepo.Logging, configuration config.CustomConfiguration) NotifierClient {
	return &Impl{
		Logging:             logging,
		CustomConfiguration: configuration,
	}
}

func (i *Impl) Setup(clientIdentifier string, url string) error {
	i.clientIdentifier = clientIdentifier
	i.url = url

	client, err := auresthttpclient.New(0, nil, nil)
	if err != nil {
		return err
	}
	aurestclientprometheus.InstrumentHttpClient(client)

	logWrapper := aurestlogging.New(client)

	circuitBreakerWrapper := aurestbreaker.New(
		logWrapper,
		fmt.Sprintf("notifier-%s-client", clientIdentifier),
		100,
		5*time.Minute,
		60*time.Second,
		// includes possible retries, once the context is cancelled further requests will fail directly
		15*time.Second,
	)
	aurestbreakerprometheus.InstrumentCircuitBreakerClient(circuitBreakerWrapper)

	// allow tests to pre-populate
	if i.Client == nil {
		i.Client = circuitBreakerWrapper
	}

	return nil
}

func (i *Impl) Send(ctx context.Context, notification openapi.Notification) {
	var responseData *[]byte
	responseDto := &aurestclientapi.ParsedResponse{
		Body: &responseData,
	}
	err := i.Client.Perform(ctx, http.MethodPost, i.url, notification, responseDto)
	if err != nil {
		i.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("failure in downstream notifier %s: %s", i.clientIdentifier, err.Error())
		return
	}
	if responseData != nil {
		i.Logging.Logger().Ctx(ctx).Info().Printf("got response result in downstream notifier %s %s", i.clientIdentifier, string(*responseData))
	}
	if responseDto.Status != http.StatusNoContent {
		i.Logging.Logger().Ctx(ctx).Warn().Printf("unexpected response status in downstream notifier %s: %d", i.clientIdentifier, responseDto.Status)
	}
}
