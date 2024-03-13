package updater

import (
	"context"
	"errors"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/errors/nochangeserror"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/repository/notifier"
	"github.com/Interhyp/metadata-service/internal/types"
	"github.com/StephanHCB/go-backend-service-common/api/apierrors"
)

// --- business logic ---

func (s *Impl) WriteRepository(ctx context.Context, key string, repository openapi.RepositoryDto) (openapi.RepositoryDto, error) {
	result := repository
	err := s.WithMetadataLock(ctx, func(subCtx context.Context) error {
		current, err := s.Cache.GetRepository(ctx, key)
		if err == nil && current.Owner != repository.Owner {

			allowed, err := s.CanMoveOrDeleteRepository(subCtx, key)
			if err != nil {
				return err
			}
			if !allowed {
				s.Logging.Logger().Ctx(ctx).Info().Printf("tried to move repository %v, which is still referenced by its service", key)
				return apierrors.NewConflictError("repository.conflict.referenced", "this repository is being referenced in a service, you cannot change its owner directly - you can change the owner of the service and this will move it along", nil, s.Timestamp.Now())
			}

			repositoryWritten, err := s.Mapper.WriteRepositoryWithChangedOwner(subCtx, key, repository)
			if err != nil {
				if nochangeserror.Is(err) {
					// there were no actual changes, this is acceptable
					result.JiraIssue = "" // cannot know, could be multiple issues for the affected files
					return nil
				}
				return err
			}
			result = repositoryWritten
		} else {
			repositoryWritten, err := s.Mapper.WriteRepository(subCtx, key, repository)
			if err != nil {
				if nochangeserror.Is(err) {
					// there were no actual changes, this is acceptable
					result.JiraIssue = "" // cannot know
					return nil
				}
				return err
			}
			result = repositoryWritten
		}

		s.fireAndForgetKafkaNotification(subCtx, s.repositoryKafkaEvent(key, result.TimeStamp, result.CommitHash))

		// cache update
		if err := s.updateRepositories(subCtx); err != nil {
			return err
		}

		return nil
	})
	return result, err
}

func (s *Impl) DeleteRepository(ctx context.Context, key string, deletionInfo openapi.DeletionDto) error {
	return s.WithMetadataLock(ctx, func(subCtx context.Context) error {
		repositoryWritten, err := s.Mapper.DeleteRepository(subCtx, key, deletionInfo.JiraIssue)
		if err != nil {
			if nochangeserror.Is(err) {
				// there were no actual changes, this is acceptable
				return nil
			}
			return err
		}

		s.fireAndForgetKafkaNotification(subCtx, s.repositoryKafkaEvent(key, repositoryWritten.TimeStamp, repositoryWritten.CommitHash))

		// cache update
		err = s.updateRepositories(subCtx)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *Impl) repositoryKafkaEvent(key string, timeStamp string, commitHash string) repository.UpdateEvent {
	return repository.UpdateEvent{
		Affected: repository.EventAffects{
			OwnerAliases:   []string{},
			ServiceNames:   []string{},
			RepositoryKeys: []string{key},
		},
		TimeStamp:  timeStamp,
		CommitHash: commitHash,
	}
}

func (s *Impl) updateRepositories(ctx context.Context) error {
	s.Logging.Logger().Ctx(ctx).Info().Print("updating repositories")

	ts := timeStamp(s.Timestamp.Now())

	repositoryKeysMap, err := s.decideRepositoriesToAddUpdateOrRemove(ctx)
	if err != nil {
		s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Print("failed to obtain repositories - skipping update this round")
		s.totalErrorCounter.Inc()
		return err
	} else {
		err = s.updateIndividualRepositories(ctx, repositoryKeysMap)
		if err != nil {
			s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Print("failed to update repositories - skipping update this round")
			return err
		} else {
			s.Logging.Logger().Ctx(ctx).Debug().Print("successfully updated repositories")
		}
	}

	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			s.Logging.Logger().Ctx(ctx).Warn().Print("timeout while updating repositories")
			return err
		}
	}

	s.Cache.SetRepositoryListTimestamp(ctx, ts)

	return nil
}

func (s *Impl) decideRepositoriesToAddUpdateOrRemove(ctx context.Context) (map[string]int8, error) {
	cachedRepositoryKeys, err := s.Cache.GetSortedRepositoryKeys(ctx)
	if err != nil {
		return nil, err
	}

	currentRepositoryKeys, err := s.Mapper.GetSortedRepositoryKeys(ctx)
	if err != nil {
		return nil, err
	}

	repositoryKeysMap := make(map[string]int8, len(currentRepositoryKeys)+len(cachedRepositoryKeys))
	for _, key := range cachedRepositoryKeys {
		s.Logging.Logger().Ctx(ctx).Debug().Printf("repository %s = removeExisting (unless overwritten)", key)
		repositoryKeysMap[key] = removeExisting
	}
	for _, key := range currentRepositoryKeys {
		_, ok := repositoryKeysMap[key]
		if !ok {
			s.Logging.Logger().Ctx(ctx).Debug().Printf("repository %s = addNew", key)
			repositoryKeysMap[key] = addNew
		} else {
			s.Logging.Logger().Ctx(ctx).Debug().Printf("repository %s = updateExisting", key)
			repositoryKeysMap[key] = updateExisting
		}
	}
	return repositoryKeysMap, nil
}

func (s *Impl) updateIndividualRepositories(ctx context.Context, repositoryKeysMap map[string]int8) error {
	var firstError error = nil
	for key, activity := range repositoryKeysMap {
		if activity == removeExisting {
			s.Logging.Logger().Ctx(ctx).Info().Printf("repository %s is no longer current, removing it from the cache", key)
			s.Cache.DeleteRepository(ctx, key)
			s.Notifier.PublishDeletion(ctx, key, types.RepositoryPayload)
		} else {
			isNew := activity == addNew
			err := s.updateIndividualRepository(ctx, key, isNew)
			if err != nil {
				if firstError == nil {
					firstError = err
				}
				if isContextCancelledOrTimeout(ctx) {
					// no use continuing the loop, everything will fail at this point
					return firstError
				}
			}
		}
	}
	return firstError
}

func (s *Impl) RefreshRepository(ctx context.Context, key string) error {
	repository, err := s.Mapper.GetRepository(ctx, key)
	if err != nil {
		return err
	}

	s.Cache.PutRepository(ctx, key, repository)
	s.Logging.Logger().Ctx(ctx).Debug().Printf("repository %s updated in cache per request", key)

	return nil
}

func (s *Impl) updateIndividualRepository(ctx context.Context, key string, isNew bool) error {
	repository, err := s.Mapper.GetRepository(ctx, key)
	if err != nil {
		if isNew {
			s.Logging.Logger().Ctx(ctx).Warn().Printf("failed to get initial info for repository %s from metadata - repository will NOT be present until next run: %s", key, err.Error())
		} else {
			s.Logging.Logger().Ctx(ctx).Warn().Printf("failed to get updated info for repository %s from metadata - repository may be outdated until next run: %s", key, err.Error())
		}
		s.totalErrorCounter.Inc()
		s.repoErrorCounter.WithLabelValues(key).Inc()
		return err
	} else {
		cached, cacheErr := s.Cache.GetRepository(ctx, key)
		s.Cache.PutRepository(ctx, key, repository)
		if isNew {
			err = s.Notifier.PublishCreation(ctx, key, notifier.AsPayload(repository))
			if err != nil {
				s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("error publishing creation of repository %s", key)
			}
			s.Logging.Logger().Ctx(ctx).Info().Printf("new repository %s added to cache", key)
		} else {
			if cacheErr == nil && !equalExceptCacheInfo(cached, repository) {
				err = s.Notifier.PublishModification(ctx, key, notifier.AsPayload(repository))
				if err != nil {
					s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("error publishing modification of repository %s", key)
				}
			}

			s.Logging.Logger().Ctx(ctx).Debug().Printf("repository %s updated in cache", key)
		}
	}
	return nil
}

func (s *Impl) CanMoveOrDeleteRepository(ctx context.Context, key string) (bool, error) {
	names, err := s.Cache.GetSortedServiceNames(ctx)
	if err != nil {
		return false, err
	}
	for _, name := range names {
		svc, _ := s.Cache.GetService(ctx, name)
		for _, candidateKey := range svc.Repositories {
			if key == candidateKey {
				return false, nil
			}
		}
	}
	return true, nil
}
