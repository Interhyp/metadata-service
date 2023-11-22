package cache

import (
	"context"
	"github.com/Interhyp/metadata-service/api"
)

const repositoryWhat = "repository"

func (s *Impl) SetRepositoryListTimestamp(ctx context.Context, timestamp string) error {
	return s.setTimestamp(ctx, repositoryWhat, repositoryTimestampKey, timestamp)
}

func (s *Impl) GetRepositoryListTimestamp(ctx context.Context) (string, error) {
	return s.getTimestamp(ctx, repositoryWhat, repositoryTimestampKey)
}

func (s *Impl) GetSortedRepositoryKeys(ctx context.Context) ([]string, error) {
	return getSortedKeys(ctx, repositoryWhat, s, s.RepositoryCache)
}

func (s *Impl) GetRepository(ctx context.Context, key string) (openapi.RepositoryDto, error) {
	return getEntry(ctx, repositoryWhat, s, s.RepositoryCache, key)
}

func (s *Impl) PutRepository(ctx context.Context, key string, entry openapi.RepositoryDto) error {
	return putEntry(ctx, repositoryWhat, s, s.RepositoryCache, key, entry)
}

func (s *Impl) DeleteRepository(ctx context.Context, key string) error {
	return removeEntry(ctx, repositoryWhat, s, s.RepositoryCache, key)
}
