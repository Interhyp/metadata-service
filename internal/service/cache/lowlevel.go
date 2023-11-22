package cache

import (
	"context"
	"fmt"
	libcache "github.com/Roshick/go-autumn-synchronisation/pkg/aucache"
	"github.com/StephanHCB/go-backend-service-common/api/apierrors"
	"sort"
)

const (
	ownerTimestampKey      = "ownertimestamp"
	serviceTimestampKey    = "servicetimestamp"
	repositoryTimestampKey = "repositorytimestamp"
)

var notFoundTimestamp = "1970-01-01T00:00:00Z"

func (s *Impl) setTimestamp(ctx context.Context, what string, key string, timestamp string) error {
	err := s.TimestampCache.Set(ctx, key, timestamp, cacheRetention)
	if err != nil {
		messageKey := fmt.Sprintf("cache.%s.error", what)
		details := fmt.Sprintf("error writing %s timestamp to cache", what)
		s.Logging.Logger().Ctx(ctx).Info().WithErr(err).Printf("%s: %s", details, err.Error())
		return apierrors.NewBadGatewayError(messageKey, details, err, s.Timestamp.Now())
	}
	return nil
}

func (s *Impl) getTimestamp(ctx context.Context, what string, key string) (string, error) {
	valPtr, err := s.TimestampCache.Get(ctx, key)
	if err != nil {
		messageKey := fmt.Sprintf("cache.%s.error", what)
		details := fmt.Sprintf("error reading %s timestamp from cache", what)
		s.Logging.Logger().Ctx(ctx).Info().Printf("%s: %s", details, err.Error())
		return notFoundTimestamp, apierrors.NewBadGatewayError(messageKey, details, err, s.Timestamp.Now())
	}
	if valPtr == nil {
		return notFoundTimestamp, err
	}
	return *valPtr, nil
}

func getSortedKeys[E any](ctx context.Context, what string, s *Impl, cache libcache.Cache[E]) ([]string, error) {
	keys, err := cache.Keys(ctx)
	if err != nil {
		messageKey := fmt.Sprintf("cache.%s.error", what)
		details := fmt.Sprintf("error reading %s keys from cache", what)
		s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("%s: %s", details, err.Error())
		return []string{}, apierrors.NewBadGatewayError(messageKey, details, err, s.Timestamp.Now())
	}
	sort.Strings(keys)
	return keys, nil
}

func getEntry[E any](ctx context.Context, what string, s *Impl, cache libcache.Cache[E], key string) (E, error) {
	copiedEntryPtr, err := cache.Get(ctx, key)
	if err != nil {
		var empty E
		messageKey := fmt.Sprintf("cache.%s.error", what)
		details := fmt.Sprintf("error reading %s %s from cache", what, key)
		s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("%s: %s", details, err.Error())
		return empty, apierrors.NewBadGatewayError(messageKey, details, err, s.Timestamp.Now())
	}
	if copiedEntryPtr == nil {
		var empty E
		messageKey := fmt.Sprintf("%s.notfound", what)
		details := fmt.Sprintf("%s %s not found", what, key)
		s.Logging.Logger().Ctx(ctx).Info().Print(details)
		return empty, apierrors.NewNotFoundError(messageKey, details, nil, s.Timestamp.Now())
	} else {
		return *copiedEntryPtr, nil
	}
}

func putEntry[E any](ctx context.Context, what string, s *Impl, cache libcache.Cache[E], key string, entry E) error {
	err := cache.Set(ctx, key, entry, cacheRetention)
	if err != nil {
		messageKey := fmt.Sprintf("cache.%s.error", what)
		details := fmt.Sprintf("error writing %s %s to cache", what, key)
		s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("%s: %s", details, err.Error())
		return apierrors.NewBadGatewayError(messageKey, details, err, s.Timestamp.Now())
	}
	return nil
}

func removeEntry[E any](ctx context.Context, what string, s *Impl, cache libcache.Cache[E], key string) error {
	err := cache.Remove(ctx, key)
	if err != nil {
		messageKey := fmt.Sprintf("cache.%s.error", what)
		details := fmt.Sprintf("error removing %s %s from cache", what, key)
		s.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("%s: %s", details, err.Error())
		return apierrors.NewBadGatewayError(messageKey, details, err, s.Timestamp.Now())
	}
	return nil
}
