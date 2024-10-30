package acceptance

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	libconfig "github.com/Interhyp/go-backend-service-common/repository/config"
	"github.com/Interhyp/go-backend-service-common/repository/logging"
	"github.com/Interhyp/go-backend-service-common/repository/timestamp"
	"github.com/Interhyp/go-backend-service-common/web/middleware/security"
	bitbucketclient "github.com/Interhyp/metadata-service/internal/client/bitbucket"
	githubclient "github.com/Interhyp/metadata-service/internal/client/github"
	"github.com/Interhyp/metadata-service/internal/repository/config"
	"github.com/Interhyp/metadata-service/internal/repository/notifier"
	"github.com/Interhyp/metadata-service/internal/service/trigger"
	"github.com/Interhyp/metadata-service/internal/service/vcswebhookshandler"
	"github.com/Interhyp/metadata-service/internal/web/app"
	"github.com/Interhyp/metadata-service/internal/web/server"
	"github.com/Interhyp/metadata-service/test/mock/idpmock"
	"github.com/Interhyp/metadata-service/test/mock/kafkamock"
	"github.com/Interhyp/metadata-service/test/mock/metadatamock"
	"github.com/Interhyp/metadata-service/test/mock/notifiermock"
	"github.com/Interhyp/metadata-service/test/mock/vaultmock"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
	aurestcapture "github.com/StephanHCB/go-autumn-restclient/implementation/capture"
	aurestplayback "github.com/StephanHCB/go-autumn-restclient/implementation/playback"
	aurestrecorder "github.com/StephanHCB/go-autumn-restclient/implementation/recorder"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"
)

// placing these here because they are package global

type ApplicationWithMocksImpl struct {
	app.ApplicationImpl
}

var (
	ts *httptest.Server

	configImpl       *libconfig.ConfigImpl
	customConfigImpl *config.CustomConfigImpl
	loggingImpl      *logging.LoggingImpl
	vaultImpl        *vaultmock.VaultImpl
	metadataImpl     *metadatamock.Impl
	kafkaImpl        *kafkamock.Impl
	idpImpl          *idpmock.Impl
	notifierImpl     *notifier.Impl

	application *ApplicationWithMocksImpl
	appCtx      context.Context
)

const validConfigurationPath = "../resources/valid-config.yaml"

func fakeNow() time.Time {
	return time.Date(2022, 11, 6, 18, 14, 10, 0, time.UTC)
}

func ConstructFilenameV4WithBody(method string, requestUrl string, body interface{}) (string, error) {
	parsedUrl, err := url.Parse(requestUrl)
	if err != nil {
		return "", err
	}

	m := strings.ToLower(method)
	p := url.QueryEscape(parsedUrl.EscapedPath())
	if len(p) > 120 {
		p = string([]byte(p)[:120])
	}
	p = strings.ReplaceAll(p, "%2F", "-")
	p = strings.TrimLeft(p, "-")
	p = strings.TrimRight(p, "-")

	if body != nil {
		parsedUrl.RawQuery = fmt.Sprintf("%v", body)
	} else if parsedUrl.RawQuery != "" {
		parsedUrl.RawQuery = fmt.Sprintf("%v", parsedUrl.Query())
	}

	// we have to ensure the filenames don't get too long. git for windows only supports 260 character paths
	md5sumOverQuery := md5.Sum([]byte(parsedUrl.Query().Encode()))
	q := hex.EncodeToString(md5sumOverQuery[:])
	q = q[:8]

	filename := fmt.Sprintf("request_%s_%s_%s.json", m, p, q)
	return filename, nil
}

func (a *ApplicationWithMocksImpl) Create() error {
	a.ConstructConfigLoggingVaultTimestamp_ForTesting()

	// prefill mocks as overrides
	a.Metadata = metadataImpl
	a.Kafka = kafkaImpl
	a.IdentityProvider = idpImpl

	// now can use normal construct functions, they respect the prefilled mocks
	if err := a.ConstructRepositories(); err != nil {
		return err
	}

	opts := aurestplayback.PlaybackOptions{
		ConstructFilenameCandidates: []aurestrecorder.ConstructFilenameFunction{ConstructFilenameV4WithBody},
	}
	bitbucketPlayback := aurestplayback.New("../resources/recordings/bitbucket", opts)
	bitbucketCapture := aurestcapture.New(bitbucketPlayback)
	bitbucketClient, _ := bitbucketclient.NewClient("localhost", "access-token")
	bitbucketClient.Client = bitbucketCapture

	githubPlayback := aurestplayback.New("../resources/recordings/github", opts)
	githubCapture := aurestcapture.NewRoundTripper(githubPlayback)
	client := http.Client{Transport: githubCapture}
	githubClient, _ := githubclient.NewClient(&client, "access-token")

	vcsPlatforms := make(map[string]vcswebhookshandler.VCSPlatform)
	vcsPlatforms["bitbucket_datacenter"] = vcswebhookshandler.VCSPlatform{
		Platform: 0,
		VCS:      bitbucketclient.New(bitbucketClient, nil),
	}
	vcsPlatforms["github"] = vcswebhookshandler.VCSPlatform{
		Platform: 1,
		VCS:      githubclient.New(githubClient, nil),
	}
	a.VCSPlatforms = &vcsPlatforms

	if err := a.ConstructServices(); err != nil {
		return err
	}
	if err := a.ConstructControllers(); err != nil {
		return err
	}

	return nil
}

func (a *ApplicationWithMocksImpl) ConstructConfigLoggingVaultTimestamp_ForTesting() {
	// construct and set up config, logging, vault, timestamp
	a.Config = configImpl
	a.CustomConfig = customConfigImpl
	a.Logging = loggingImpl
	a.Vault = vaultImpl
	a.Timestamp = timestamp.NewNoAcorn(fakeNow)
}

func (a *ApplicationWithMocksImpl) Teardown() {
	// reverse order (must ensure correct order yourself, but most components will not have a teardown method)
	a.Trigger.Teardown()
	a.Kafka.Teardown()
	a.Metadata.Teardown()
}

// use a special configuration and wire in mocks for most repositories
func tstSetup(configPath string) error {
	application = &ApplicationWithMocksImpl{}
	err := tstSetupConfig(configPath)
	if err != nil {
		return err
	}
	tstSetupLogging()

	vaultImpl = vaultmock.New().(*vaultmock.VaultImpl)
	metadataImpl = metadatamock.New().(*metadatamock.Impl)
	kafkaImpl = kafkamock.New().(*kafkamock.Impl)
	idpImpl = idpmock.New().(*idpmock.Impl)

	metadataImpl.Now = fakeNow

	err = application.Create()
	if err != nil {
		return err
	}

	application.Trigger.(*trigger.Impl).SkipStart = true // do not start cron job

	notifierImpl = application.Notifier.(*notifier.Impl)
	notifierImpl.SkipAsync = true

	security.Now = fakeNow

	for identifier, _ := range notifierImpl.Clients {
		notifierImpl.Clients[identifier] = &notifiermock.NotifierClientMock{SentNotifications: make([]string, 0)}
	}

	tstSetupHttpTestServer()
	return nil
}

func tstSetupConfig(configPath string) error {
	impl, cImpl := config.New()
	configImpl = impl.(*libconfig.ConfigImpl)
	customConfigImpl = cImpl.(*config.CustomConfigImpl)
	auconfigenv.LocalConfigFileName = configPath
	err := configImpl.Read()
	if err != nil {
		return err
	}
	// do not wish to validate configuration, so setting parsed values directly
	configImpl.ObtainPredefinedValues()
	configImpl.CustomConfiguration.Obtain(auconfigenv.Get)

	customConfigImpl := configImpl.CustomConfiguration.(*config.CustomConfigImpl)
	// and can override configuration values here
	customConfigImpl.VUpdateJobTimeoutSeconds = 1
	return nil
}

func tstSetupLogging() {
	loggingImpl = logging.New().(*logging.LoggingImpl)
	loggingImpl.SetupForTesting()
	appCtx = log.Logger.WithContext(context.Background())
	configImpl.Logging = loggingImpl
	loggingImpl.Configuration = configImpl
}

func tstSetupHttpTestServer() {
	application.Server.WireUp(appCtx)
	ts = httptest.NewServer(application.Server.(*server.Impl).Router)
}

func tstShutdown() {
	ts.Close()
}

func tstReset() {
	metadataImpl.Reset()
	kafkaImpl.Reset()
	for _, client := range notifierImpl.Clients {
		client.(*notifiermock.NotifierClientMock).Reset()
	}
}
