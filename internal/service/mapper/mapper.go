package mapper

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/repository"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration repository.CustomConfiguration
	Logging             librepo.Logging
	Metadata            repository.Metadata

	muOwnerCaches        sync.Mutex
	serviceOwnerCache    map[string]string
	repositoryOwnerCache map[string]string

	OwnerRegex *regexp.Regexp
	Now        func() time.Time
}

func (s *Impl) Setup(_ context.Context) error {
	ownerRegex, err := regexp.Compile(s.CustomConfiguration.OwnerRegex())
	if err != nil {
		return err
	}

	s.OwnerRegex = ownerRegex

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
