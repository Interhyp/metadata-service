package mapper

import (
	"context"
	"fmt"
	"github.com/Interhyp/metadata-service/acorns/errors/nochangeserror"
	openapi "github.com/Interhyp/metadata-service/api/v1"
	"github.com/StephanHCB/go-backend-service-common/web/middleware/requestid"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"time"
)

type Dtos interface {
	openapi.OwnerDto | openapi.ServiceDto | openapi.RepositoryDto
}

type PatchDtos interface {
	openapi.OwnerPatchDto | openapi.ServicePatchDto | openapi.RepositoryPatchDto
}

func SetCommitHash(dto interface{}, commitHash string) {
	if i, ok := dto.(*openapi.OwnerDto); ok {
		i.CommitHash = commitHash
	} else if i, ok := dto.(*openapi.ServiceDto); ok {
		i.CommitHash = commitHash
	} else if i, ok := dto.(*openapi.RepositoryDto); ok {
		i.CommitHash = commitHash
	} else if i, ok := dto.(*openapi.OwnerPatchDto); ok {
		i.CommitHash = commitHash
	} else if i, ok := dto.(*openapi.ServicePatchDto); ok {
		i.CommitHash = commitHash
	} else if i, ok := dto.(*openapi.RepositoryPatchDto); ok {
		i.CommitHash = commitHash
	}
}

func SetTimeStamp(dto interface{}, rawTimeStamp time.Time) {
	if i, ok := dto.(*openapi.OwnerDto); ok {
		i.TimeStamp = timeStamp(rawTimeStamp)
	} else if i, ok := dto.(*openapi.ServiceDto); ok {
		i.TimeStamp = timeStamp(rawTimeStamp)
	} else if i, ok := dto.(*openapi.RepositoryDto); ok {
		i.TimeStamp = timeStamp(rawTimeStamp)
	} else if i, ok := dto.(*openapi.OwnerPatchDto); ok {
		i.TimeStamp = timeStamp(rawTimeStamp)
	} else if i, ok := dto.(*openapi.ServicePatchDto); ok {
		i.TimeStamp = timeStamp(rawTimeStamp)
	} else if i, ok := dto.(*openapi.RepositoryPatchDto); ok {
		i.TimeStamp = timeStamp(rawTimeStamp)
	}
}

func SetJiraIssue(dto interface{}, commitMessage string) {
	if i, ok := dto.(*openapi.OwnerDto); ok {
		i.JiraIssue = jiraIssue(commitMessage)
	} else if i, ok := dto.(*openapi.ServiceDto); ok {
		i.JiraIssue = jiraIssue(commitMessage)
	} else if i, ok := dto.(*openapi.RepositoryDto); ok {
		i.JiraIssue = jiraIssue(commitMessage)
	} else if i, ok := dto.(*openapi.OwnerPatchDto); ok {
		i.JiraIssue = jiraIssue(commitMessage)
	} else if i, ok := dto.(*openapi.ServicePatchDto); ok {
		i.JiraIssue = jiraIssue(commitMessage)
	} else if i, ok := dto.(*openapi.RepositoryPatchDto); ok {
		i.JiraIssue = jiraIssue(commitMessage)
	}
}

func GetT[T Dtos](_ context.Context, s *Impl, resultPtr *T, fullPath string) error {
	yamlBytes, commitInfo, err := s.Metadata.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read %s from metadata: %s", fullPath, err.Error())
	}

	err = yaml.Unmarshal(yamlBytes, resultPtr)
	if err != nil {
		return fmt.Errorf("failed to parse %s as yaml from metadata: %s", fullPath, err.Error())
	}

	SetCommitHash(resultPtr, commitInfo.CommitHash)
	SetTimeStamp(resultPtr, commitInfo.TimeStamp)
	SetJiraIssue(resultPtr, commitInfo.Message)
	return nil
}

func (s *Impl) resetLocalClone(ctx context.Context) {
	reqId := ctx.Value(requestid.RequestIDKey)
	if reqId == nil {
		reqId = requestid.NewRequestID()
	}
	newCtx := context.WithValue(context.Background(), requestid.RequestIDKey, reqId)

	logger := log.Ctx(ctx)
	newCtx = logger.WithContext(newCtx)

	err := s.Metadata.Clone(newCtx)
	if err != nil {
		s.Logging.Logger().Ctx(newCtx).Error().WithErr(err).Print("failed to repair local clone - continuing")
	}
}

func WriteT[T Dtos](ctx context.Context, s *Impl, resultPtr *T, path string, fileNameNoPath string, description string, jiraIssue string) error {
	fileName := path + "/" + fileNameNoPath

	yamlBytes, err := yaml.Marshal(*resultPtr)
	if err != nil {
		return err
	}

	err = s.Metadata.MkdirAll(path)
	if err != nil {
		s.resetLocalClone(ctx)
		return err
	}

	err = s.Metadata.WriteFile(fileName, yamlBytes)
	if err != nil {
		s.resetLocalClone(ctx)
		return err
	}

	message := fmt.Sprintf("%s: update %s", jiraIssue, description)
	commitInfo, err := s.Metadata.Commit(ctx, message)
	if err != nil {
		if !nochangeserror.Is(err) {
			// empty commits need no re-clone
			s.resetLocalClone(ctx)
		}
		SetJiraIssue(resultPtr, "")
		return err
	}

	SetCommitHash(resultPtr, commitInfo.CommitHash)
	SetTimeStamp(resultPtr, commitInfo.TimeStamp)
	SetJiraIssue(resultPtr, commitInfo.Message)

	err = s.Metadata.Push(ctx)
	if err != nil {
		s.resetLocalClone(ctx)
		return err
	}

	return nil
}

func DeleteT[T PatchDtos](ctx context.Context, s *Impl, resultPtr *T, fullPath string, description string, jiraIssue string) error {
	err := s.Metadata.DeleteFile(fullPath)
	if err != nil {
		s.resetLocalClone(ctx)
		return err
	}

	message := fmt.Sprintf("%s: delete %s", jiraIssue, description)
	commitInfo, err := s.Metadata.Commit(ctx, message)
	if err != nil {
		if !nochangeserror.Is(err) {
			// empty commits need no re-clone
			s.resetLocalClone(ctx)
		}
		return err
	}

	SetCommitHash(resultPtr, commitInfo.CommitHash)
	SetTimeStamp(resultPtr, commitInfo.TimeStamp)
	SetJiraIssue(resultPtr, commitInfo.Message)

	err = s.Metadata.Push(ctx)
	if err != nil {
		s.resetLocalClone(ctx)
		return err
	}

	return nil
}

func Move(ctx context.Context, s *Impl, v interface{}, oldFullPath string, newPath string, newFileNameNoPath string) error {
	err := s.Metadata.DeleteFile(oldFullPath)
	if err != nil {
		return err
	}

	err = s.Metadata.MkdirAll(newPath)
	if err != nil {
		return err
	}

	yamlBytes, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	err = s.Metadata.WriteFile(newPath+"/"+newFileNameNoPath, yamlBytes)
	if err != nil {
		return err
	}

	return nil
}
