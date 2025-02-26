package mapper

import (
	"context"
	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/service"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	"strings"
	"sync"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Metadata            repository.Metadata
	Timestamp           librepo.Timestamp

	muOwnerCaches        sync.Mutex
	serviceOwnerCache    map[string]string
	repositoryOwnerCache map[string]string
}

func New(
	configuration librepo.Configuration,
	customConfig config.CustomConfiguration,
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	metadata repository.Metadata,
) service.Mapper {
	return &Impl{
		Configuration:       configuration,
		CustomConfiguration: customConfig,
		Logging:             logging,
		Timestamp:           timestamp,
		Metadata:            metadata,
	}
}

func (s *Impl) IsMapper() bool {
	return true
}

func (s *Impl) Setup() error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	err := s.SetupMapper(ctx)
	if err != nil {
		s.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up mapper. BAILING OUT.")
		return err
	}

	s.Logging.Logger().Ctx(ctx).Info().Print("successfully set up mapper")
	return nil
}

func (s *Impl) SetupMapper(_ context.Context) error {
	s.serviceOwnerCache = make(map[string]string)
	s.repositoryOwnerCache = make(map[string]string)

	return nil
}

func (s *Impl) RefreshMetadata(ctx context.Context) ([]repository.UpdateEvent, error) {
	events := make([]repository.UpdateEvent, 0)

	err := s.Metadata.Pull(ctx)
	if err != nil {
		return events, err
	}

	newCommits := s.Metadata.NewPulledCommits()
	for _, commitInfo := range newCommits {
		event := repository.UpdateEvent{
			Affected: repository.EventAffects{
				OwnerAliases:   ownerAliasesFromCommitInfo(commitInfo),
				ServiceNames:   serviceNamesFromCommitInfo(commitInfo),
				RepositoryKeys: repoKeysFromCommitInfo(commitInfo),
			},
			TimeStamp:  timeStamp(commitInfo.TimeStamp),
			CommitHash: commitInfo.CommitHash,
		}
		events = append(events, event)
	}
	return events, nil
}

func ownerAliasesFromCommitInfo(commitInfo repository.CommitInfo) []string {
	result := make([]string, 0)
	for _, path := range commitInfo.FilesChanged {
		components := strings.Split(path, "/")
		if len(components) == 3 && components[0] == "owners" && components[2] == "owner.info.yaml" {
			result = append(result, components[1])
		}
	}
	return result
}

func serviceNamesFromCommitInfo(commitInfo repository.CommitInfo) []string {
	result := make([]string, 0)
	for _, path := range commitInfo.FilesChanged {
		components := strings.Split(path, "/")
		if len(components) == 4 && components[0] == "owners" && components[2] == "services" {
			parts := strings.Split(components[3], ".")
			if len(parts) == 2 {
				result = append(result, parts[0])
			}
		}
	}
	return result
}

func repoKeysFromCommitInfo(commitInfo repository.CommitInfo) []string {
	result := make([]string, 0)
	for _, path := range commitInfo.FilesChanged {
		components := strings.Split(path, "/")
		if len(components) == 4 && components[0] == "owners" && components[2] == "repositories" {
			parts := strings.Split(components[3], ".")
			if len(parts) == 3 {
				result = append(result, parts[0]+"."+parts[1])
			}
		}
	}
	return result
}

func (s *Impl) ContainsNewInformation(_ context.Context, event repository.UpdateEvent) bool {
	return !s.Metadata.IsCommitKnown(event.CommitHash)
}
