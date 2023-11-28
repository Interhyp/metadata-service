package cache

import (
	"context"
	"github.com/Interhyp/metadata-service/api"
)

const ownerWhat = "owner"

func (s *Impl) SetOwnerListTimestamp(ctx context.Context, timestamp string) error {
	return s.setTimestamp(ctx, ownerWhat, ownerTimestampKey, timestamp)
}

func (s *Impl) GetOwnerListTimestamp(ctx context.Context) (string, error) {
	return s.getTimestamp(ctx, ownerWhat, ownerTimestampKey)
}

func (s *Impl) GetSortedOwnerAliases(ctx context.Context) ([]string, error) {
	return getSortedKeys(ctx, ownerWhat, s, s.OwnerCache)
}

func (s *Impl) GetOwner(ctx context.Context, alias string) (openapi.OwnerDto, error) {
	return getEntry(ctx, ownerWhat, s, s.OwnerCache, alias)
}

func (s *Impl) PutOwner(ctx context.Context, alias string, entry openapi.OwnerDto) error {
	return putEntry(ctx, ownerWhat, s, s.OwnerCache, alias, entry)
}

func (s *Impl) DeleteOwner(ctx context.Context, alias string) error {
	return removeEntry(ctx, ownerWhat, s, s.OwnerCache, alias)
}
