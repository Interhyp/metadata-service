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

func (s *Impl) WriteService(ctx context.Context, serviceName string, service openapi.ServiceDto) (openapi.ServiceDto, error) {
	result := service
	err := s.WithMetadataLock(ctx, func(subCtx context.Context) error {
		current, err := s.Cache.GetService(ctx, serviceName)
		if err == nil && current.Owner != service.Owner {
			serviceWritten, err := s.Mapper.WriteServiceWithChangedOwner(subCtx, serviceName, service)
			if err != nil {
				if nochangeserror.Is(err) {
					// there were no actual changes, this is acceptable
					result.JiraIssue = "" // cannot know, could be multiple issues for the affected files
					return nil
				}
				return err
			}
			result = serviceWritten

			s.fireAndForgetKafkaNotification(subCtx, s.serviceAndReposKafkaEvent(serviceName, service.Repositories, serviceWritten.TimeStamp, serviceWritten.CommitHash))

			// cache updates (incl. repositories)
			if err := s.updateServices(subCtx); err != nil {
				return err
			}

			if err := s.updateRepositories(subCtx); err != nil {
				return err
			}
		} else {
			serviceWritten, err := s.Mapper.WriteService(subCtx, serviceName, service)
			if err != nil {
				if nochangeserror.Is(err) {
					// there were no actual changes, this is acceptable
					result.JiraIssue = "" // cannot know
					return nil
				}
				return err
			}
			result = serviceWritten

			s.fireAndForgetKafkaNotification(subCtx, s.serviceKafkaEvent(serviceName, serviceWritten.TimeStamp, serviceWritten.CommitHash))

			// cache update
			if err := s.updateServices(subCtx); err != nil {
				return err
			}
		}

		return nil
	})
	return result, err
}

func (s *Impl) DeleteService(ctx context.Context, serviceName string, deletionInfo openapi.DeletionDto) error {
	return s.WithMetadataLock(ctx, func(subCtx context.Context) error {
		serviceWritten, err := s.Mapper.DeleteService(subCtx, serviceName, deletionInfo.JiraIssue)
		if err != nil {
			if nochangeserror.Is(err) {
				// there were no actual changes, this is acceptable
				return nil
			}
			return err
		}

		s.fireAndForgetKafkaNotification(subCtx, s.serviceKafkaEvent(serviceName, serviceWritten.TimeStamp, serviceWritten.CommitHash))

		// cache update
		err = s.updateServices(subCtx)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *Impl) serviceKafkaEvent(serviceName string, timeStamp string, commitHash string) repository.UpdateEvent {
	return repository.UpdateEvent{
		Affected: repository.EventAffects{
			OwnerAliases:   []string{},
			ServiceNames:   []string{serviceName},
			RepositoryKeys: []string{},
		},
		TimeStamp:  timeStamp,
		CommitHash: commitHash,
	}
}

func (s *Impl) serviceAndReposKafkaEvent(serviceName string, repoKeys []string, timeStamp string, commitHash string) repository.UpdateEvent {
	return repository.UpdateEvent{
		Affected: repository.EventAffects{
			OwnerAliases:   []string{},
			ServiceNames:   []string{serviceName},
			RepositoryKeys: repoKeys,
		},
		TimeStamp:  timeStamp,
		CommitHash: commitHash,
	}
}

func (s *Impl) updateServices(ctx context.Context) error {
	s.Logging.Logger().Ctx(ctx).Info().Print("updating services")

	ts := timeStamp(s.Now())

	serviceNamesMap, err := s.decideServicesToAddUpdateOrRemove(ctx)
	if err != nil {
		s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Print("failed to obtain services - skipping update this round")
		s.totalErrorCounter.Inc()
		return err
	} else {
		err = s.updateIndividualServices(ctx, serviceNamesMap)
		if err != nil {
			s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Print("failed to update services - skipping update this round")
			return err
		} else {
			s.Logging.Logger().Ctx(ctx).Debug().Print("successfully updated services")
		}
	}

	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			s.Logging.Logger().Ctx(ctx).Warn().Print("timeout while updating services")
			return err
		}
	}

	s.Cache.SetServiceListTimestamp(ctx, ts)

	return nil
}

func (s *Impl) decideServicesToAddUpdateOrRemove(ctx context.Context) (map[string]int8, error) {
	cachedServiceNames := s.Cache.GetSortedServiceNames(ctx)

	currentServiceNames, err := s.Mapper.GetSortedServiceNames(ctx)
	if err != nil {
		return nil, err
	}

	serviceNamesMap := make(map[string]int8, len(currentServiceNames)+len(cachedServiceNames))
	for _, name := range cachedServiceNames {
		s.Logging.Logger().Ctx(ctx).Debug().Printf("service %s = removeExisting (unless overwritten)", name)
		serviceNamesMap[name] = removeExisting
	}
	for _, name := range currentServiceNames {
		_, ok := serviceNamesMap[name]
		if !ok {
			s.Logging.Logger().Ctx(ctx).Debug().Printf("service %s = addNew", name)
			serviceNamesMap[name] = addNew
		} else {
			s.Logging.Logger().Ctx(ctx).Debug().Printf("service %s = updateExisting", name)
			serviceNamesMap[name] = updateExisting
		}
	}
	return serviceNamesMap, nil
}

func (s *Impl) updateIndividualServices(ctx context.Context, serviceNamesMap map[string]int8) error {
	var firstError error = nil
	for name, activity := range serviceNamesMap {
		if activity == removeExisting {
			s.Logging.Logger().Ctx(ctx).Info().Printf("service %s is no longer current, removing it from the cache", name)
			s.Cache.DeleteService(ctx, name)
			s.Notifier.PublishDeletion(ctx, name, types.ServicePayload)
		} else {
			isNew := activity == addNew
			err := s.updateIndividualService(ctx, name, isNew)
			if err != nil {
				if firstError == nil {
					firstError = err
				}
			}
		}
	}
	return firstError
}

func (s *Impl) RefreshService(ctx context.Context, serviceName string) error {
	service, err := s.Mapper.GetService(ctx, serviceName)
	if err != nil {
		return err
	}

	s.Cache.PutService(ctx, serviceName, service)
	s.Logging.Logger().Ctx(ctx).Debug().Printf("service %s updated in cache per request", serviceName)

	return nil
}

func (s *Impl) updateIndividualService(ctx context.Context, name string, isNew bool) error {
	service, err := s.Mapper.GetService(ctx, name)
	if err != nil {
		if isNew {
			s.Logging.Logger().Ctx(ctx).Warn().Printf("failed to get initial info for service %s from metadata - service will NOT be present until next run: %s", name, err.Error())
		} else {
			s.Logging.Logger().Ctx(ctx).Warn().Printf("failed to get updated info for service %s from metadata - service may be outdated until next run: %s", name, err.Error())
		}
		s.totalErrorCounter.Inc()
		s.serviceErrorCounter.WithLabelValues(name).Inc()
		return err
	} else {
		s.Cache.PutService(ctx, name, service)
		if isNew {
			err = s.Notifier.PublishCreation(ctx, name, notifier.AsPayload(service))
			if err != nil {
				s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("error publishing creation of service %s", service)
			}
			s.Logging.Logger().Ctx(ctx).Info().Printf("new service %s added to cache", name)
		} else {
			err = s.Notifier.PublishModification(ctx, name, notifier.AsPayload(service))
			if err != nil {
				s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("error publishing modification of service %s", service)
			}
			s.Logging.Logger().Ctx(ctx).Debug().Printf("service %s updated in cache", name)
		}
	}
	return nil
}
