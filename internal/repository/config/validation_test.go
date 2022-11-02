package config

import (
	"bytes"
	"context"
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/Interhyp/metadata-service/docs"
	auconfigenv "github.com/StephanHCB/go-autumn-config-env"
	goauzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	libconfig "github.com/StephanHCB/go-backend-service-common/repository/config"
	"github.com/StephanHCB/go-backend-service-common/repository/logging"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"testing"
)

const basedir = "../../../test/resources/"

func tstYamlRead(t *testing.T, filename string, expectedMsgPart string) {
	cut := New().(*libconfig.ConfigImpl)
	auconfigenv.LocalConfigFileName = basedir + filename
	err := cut.Read()
	if expectedMsgPart == "" {
		require.Nil(t, err)
	} else {
		require.NotNil(t, err)
		require.Contains(t, err.Error(), expectedMsgPart)
	}
}

func TestYamlRead_MissingFile(t *testing.T) {
	docs.Description("the local configuration file is optional")
	tstYamlRead(t, "not-there.yaml", "")
}

func TestYamlRead_InvalidSyntax(t *testing.T) {
	docs.Description("the local configuration must be correct yaml syntax")
	tstYamlRead(t, "invalid-config-syntax.yaml", "error parsing local configuration flat yaml file")
}

func tstSetupCutAndLogRecorder(t *testing.T, configfile string) (librepo.Configuration, error) {
	cut := New().(librepo.Configuration)

	// --- simulate auacornapi.Acorn Assemble phase for just the configuration

	auconfigenv.LocalConfigFileName = basedir + configfile
	err := cut.Read()
	require.Nil(t, err)

	// --- simulate auacornapi.Acorn Setup phase for just the configuration, adding a mock log recorder

	// set up log recorder
	logRecorder := logging.New().(librepo.Logging)
	goauzerolog.RecordedLogForTesting = new(bytes.Buffer)
	logRecorder.(*logging.LoggingImpl).SetupForTesting()

	cut.(*libconfig.ConfigImpl).Logging = logRecorder

	ctx := log.Logger.WithContext(context.Background())
	err = cut.Validate(ctx)

	cut.(*libconfig.ConfigImpl).ObtainPredefinedValues()
	cut.(*libconfig.ConfigImpl).CustomConfiguration.Obtain(auconfigenv.Get)

	return cut, err
}

func TestValidate_LotsOfErrors(t *testing.T) {
	docs.Description("validation of configuration values works")

	_, err := tstSetupCutAndLogRecorder(t, "invalid-config-values.yaml")

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "some configuration values failed to validate or parse. There were 20 error(s). See details above")

	actualLog := goauzerolog.RecordedLogForTesting.String()

	expectedPart := "\"message\":\"failed to validate configuration field ENVIRONMENT: must match ^(feat|"
	require.Contains(t, actualLog, expectedPart)

	expectedPart2 := "\"message\":\"failed to validate configuration field SERVER_PORT: value 122834 is out of range [1024..65535]"
	require.Contains(t, actualLog, expectedPart2)

	expectedPart3 := "METRICS_PORT: value -12387192873invalid is not a valid integer"
	require.Contains(t, actualLog, expectedPart3)

	expectedPart4 := "failed to validate configuration field KAFKA_SEED_BROKERS: must match ^(|([a-z0-9-]+.[a-z0-9-]+.[a-z]{2,3}"
	require.Contains(t, actualLog, expectedPart4)

	expectedPart5 := "failed to validate configuration field ALERT_TARGET_PREFIX: must match ^((http|https)://|)[a-z0-9-.]+.[a-z]{2,3}/$"
	require.Contains(t, actualLog, expectedPart5)

	expectedPart6 := "failed to validate configuration field ALERT_TARGET_SUFFIX: must match ^@[a-z0-9-]+.[a-z]{2,3}$"
	require.Contains(t, actualLog, expectedPart6)
}

func TestAccessors(t *testing.T) {
	docs.Description("the config accessors return the correct values")

	cut, err := tstSetupCutAndLogRecorder(t, "valid-config-unique.yaml")

	require.Nil(t, err)

	actualLog := goauzerolog.RecordedLogForTesting.String()
	require.Equal(t, "", actualLog)

	require.Equal(t, "room-service", cut.ApplicationName())
	require.Equal(t, "192.168.150.0", cut.ServerAddress())
	require.Equal(t, uint16(8081), cut.ServerPort())
	require.Equal(t, uint16(9091), cut.MetricsPort())
	require.Equal(t, "dev", cut.Environment())
	require.Equal(t, true, cut.PlainLogging())
	require.Equal(t, "localhost", cut.VaultServer())
	require.Equal(t, "", cut.VaultCertificateFile())
	require.Equal(t, "room-service/secrets", cut.VaultSecretPath())
	require.Equal(t, true, cut.LocalVault())
	require.Equal(t, "not a real token", cut.LocalVaultToken())
	require.Equal(t, "example_microservice_role_room-service_prod", cut.VaultKubernetesRole())
	require.Equal(t, "/some/thing", cut.VaultKubernetesTokenPath())
	require.Equal(t, "k8s-dev-something", cut.VaultKubernetesBackend())

	require.Equal(t, "somebody", repository.Custom(cut).BbUser())
	require.Equal(t, "Body, Some", repository.Custom(cut).GitCommitterName())
	require.Equal(t, "somebody@somewhere.com", repository.Custom(cut).GitCommitterEmail())
	require.Equal(t, "some-kafka-user", repository.Custom(cut).KafkaUser())
	require.Equal(t, "some-kafka-topic", repository.Custom(cut).KafkaTopic())
	require.Equal(t, "first-kafka-broker.domain.com:9092,second-kafka-broker.domain.com:9092", repository.Custom(cut).KafkaSeedBrokers())
	require.Equal(t, "http://keyset", repository.Custom(cut).KeySetUrl())
	require.Equal(t, "http://metadata", repository.Custom(cut).MetadataRepoUrl())
	require.Equal(t, "[a-z][0-9]+", repository.Custom(cut).OwnerRegex())
	require.Equal(t, "5", repository.Custom(cut).UpdateJobIntervalCronPart())
	require.Equal(t, uint16(30), repository.Custom(cut).UpdateJobTimeoutSeconds())
	require.Equal(t, "base/path", repository.Custom(cut).VaultSecretsBasePath())
	require.Equal(t, "kafka/feat/room-service", repository.Custom(cut).VaultKafkaSecretPath())
	require.Equal(t, "https://some-domain.com/", repository.Custom(cut).AlertTargetPrefix())
	require.Equal(t, "@some-domain.com", repository.Custom(cut).AlertTargetSuffix())
	require.EqualValues(t, []string{"someguy"}, repository.Custom(cut).AdditionalPromoters())
	require.EqualValues(t, []string{"add-my-promoters-to-every-service", "also-add-my-promoters"}, repository.Custom(cut).AdditionalPromotersFromOwners())
}
