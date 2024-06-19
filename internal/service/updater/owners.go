package updater

import (
	"context"
	"errors"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/errors/nochangeserror"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/repository/notifier"
	"github.com/Interhyp/metadata-service/internal/types"
)

// --- business logic ---

func (s *Impl) WriteOwner(ctx context.Context, ownerAlias string, owner openapi.OwnerDto) (openapi.OwnerDto, error) {
	result := owner
	err := s.WithMetadataLock(ctx, func(subCtx context.Context) error {
		ownerWritten, err := s.Mapper.WriteOwner(subCtx, ownerAlias, owner)
		if err != nil {
			if nochangeserror.Is(err) {
				// there were no actual changes, this is acceptable
				result.JiraIssue = "" // cannot know
				return nil
			}
			return err
		}
		result = ownerWritten

		s.fireAndForgetKafkaNotification(subCtx, s.ownerKafkaEvent(ownerAlias, ownerWritten.TimeStamp, ownerWritten.CommitHash))

		// cache update
		err = s.updateOwners(subCtx)
		if err != nil {
			return err
		}

		return nil
	})
	return result, err
}

func (s *Impl) DeleteOwner(ctx context.Context, ownerAlias string, deletionInfo openapi.DeletionDto) error {
	return s.WithMetadataLock(ctx, func(subCtx context.Context) error {
		ownerWritten, err := s.Mapper.DeleteOwner(subCtx, ownerAlias, deletionInfo.JiraIssue)
		if err != nil {
			if nochangeserror.Is(err) {
				// there were no actual changes, this is acceptable
				return nil
			}
			return err
		}

		s.fireAndForgetKafkaNotification(subCtx, s.ownerKafkaEvent(ownerAlias, ownerWritten.TimeStamp, ownerWritten.CommitHash))

		// cache update
		err = s.updateOwners(subCtx)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *Impl) CanDeleteOwner(ctx context.Context, ownerAlias string) bool {
	return s.Mapper.IsOwnerEmpty(ctx, ownerAlias)
}

func (s *Impl) ownerKafkaEvent(ownerAlias string, timeStamp string, commitHash string) repository.UpdateEvent {
	return repository.UpdateEvent{
		Affected: repository.EventAffects{
			OwnerAliases:   []string{ownerAlias},
			ServiceNames:   []string{},
			RepositoryKeys: []string{},
		},
		TimeStamp:  timeStamp,
		CommitHash: commitHash,
	}
}

func (s *Impl) updateOwners(ctx context.Context) error {
	s.Logging.Logger().Ctx(ctx).Info().Print("updating owners")

	ts := timeStamp(s.Timestamp.Now())

	ownerAliasesMap, err := s.decideOwnersToAddUpdateOrRemove(ctx)
	if err != nil {
		return err
	}

	err = s.updateIndividualOwners(ctx, ownerAliasesMap)
	if err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			s.Logging.Logger().Ctx(ctx).Warn().Print("timeout while updating owners")
			return err
		}
	}

	s.Cache.SetOwnerListTimestamp(ctx, ts)

	return nil
}

func (s *Impl) decideOwnersToAddUpdateOrRemove(ctx context.Context) (map[string]int8, error) {
	cachedOwnerAliases, err := s.Cache.GetSortedOwnerAliases(ctx)
	if err != nil {
		return nil, err
	}

	currentOwnerAliases, err := s.Mapper.GetSortedOwnerAliases(ctx)
	if err != nil {
		return nil, err
	}

	ownerAliasesMap := make(map[string]int8, len(currentOwnerAliases)+len(cachedOwnerAliases))
	for _, alias := range cachedOwnerAliases {
		s.Logging.Logger().Ctx(ctx).Debug().Printf("owner %s = removeExisting (unless overwritten)", alias)
		ownerAliasesMap[alias] = removeExisting
	}
	for _, alias := range currentOwnerAliases {
		_, ok := ownerAliasesMap[alias]
		if !ok {
			s.Logging.Logger().Ctx(ctx).Debug().Printf("owner %s = addNew", alias)
			ownerAliasesMap[alias] = addNew
		} else {
			s.Logging.Logger().Ctx(ctx).Debug().Printf("owner %s = updateExisting", alias)
			ownerAliasesMap[alias] = updateExisting
		}
	}
	return ownerAliasesMap, nil
}

func (s *Impl) updateIndividualOwners(ctx context.Context, ownerAliasesMap map[string]int8) error {
	var firstError error = nil
	for alias, activity := range ownerAliasesMap {
		var err error
		if activity == removeExisting {
			s.removeIndividualOwner(ctx, alias)
		} else if activity == addNew {
			err = s.addIndividualOwner(ctx, alias)
		} else {
			err = s.updateIndividualOwner(ctx, alias)
		}
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
	return firstError
}

func (s *Impl) removeIndividualOwner(ctx context.Context, alias string) {
	s.Logging.Logger().Ctx(ctx).Info().Printf("owner %s is no longer current, removing it from the cache", alias)
	s.Cache.DeleteOwner(ctx, alias)
	s.Notifier.PublishDeletion(ctx, alias, types.OwnerPayload)
}

func (s *Impl) addIndividualOwner(ctx context.Context, alias string) error {
	owner, err := s.Mapper.GetOwner(ctx, alias)
	if err != nil {
		s.Logging.Logger().Ctx(ctx).Warn().Printf("failed to get initial info for owner %s from metadata - owner will NOT be present until next run: %s", alias, err.Error())
		s.totalErrorCounter.Inc()
	} else {
		s.Cache.PutOwner(ctx, alias, owner)
		if errOnlyLog := s.Notifier.PublishCreation(ctx, alias, notifier.AsPayload(owner)); errOnlyLog != nil {
			s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("error publishing creation of owner %s", alias)
		}
		s.Logging.Logger().Ctx(ctx).Info().Printf("new owner %s added to cache", alias)
	}
	return err
}

func (s *Impl) updateIndividualOwner(ctx context.Context, alias string) error {
	owner, err := s.Mapper.GetOwner(ctx, alias)
	if err != nil {
		s.Logging.Logger().Ctx(ctx).Warn().Printf("failed to get updated info for owner %s from metadata - owner may be outdated until next run: %s", alias, err.Error())
		s.totalErrorCounter.Inc()
	} else {
		cached, cacheErr := s.Cache.GetOwner(ctx, alias)

		s.Cache.PutOwner(ctx, alias, owner)
		if cacheErr == nil && !equalExceptCacheInfo(cached, owner) {
			if errOnlyLog := s.Notifier.PublishModification(ctx, alias, notifier.AsPayload(owner)); errOnlyLog != nil {
				s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("error publishing modification of owner %s", alias)
			}
		}
		s.Logging.Logger().Ctx(ctx).Debug().Printf("owner %s updated in cache", alias)
	}
	return err
}

func isContextCancelledOrTimeout(ctx context.Context) bool {
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return true
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return true
		}
	}
	return false
}
