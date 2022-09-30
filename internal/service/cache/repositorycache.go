package cache

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/errors/nosuchrepoerror"
	openapi "github.com/Interhyp/metadata-service/api/v1"
)

func (s *Impl) SetRepositoryListTimestamp(_ context.Context, timestamp string) {
	s.RepositoryCache.SetTimestamp(timestamp)
}

func (s *Impl) GetRepositoryListTimestamp(_ context.Context) string {
	return s.RepositoryCache.GetTimestamp()
}

func (s *Impl) GetSortedRepositoryKeys(_ context.Context) []string {
	keysPtr := s.RepositoryCache.GetSortedKeys()
	return deepCopyStringSlice(*keysPtr)
}

func (s *Impl) GetRepository(ctx context.Context, key string) (openapi.RepositoryDto, error) {
	immutableRepositoryPtr := s.RepositoryCache.GetEntryRef(key)
	if immutableRepositoryPtr == nil {
		return openapi.RepositoryDto{}, nosuchrepoerror.New(ctx, key)
	} else {
		return deepCopyRepository(immutableRepositoryPtr), nil
	}
}

func (s *Impl) PutRepository(_ context.Context, key string, entry openapi.RepositoryDto) {
	var e interface{}
	e = entry

	s.RepositoryCache.UpdateEntryRef(key, &e)
}

func (s *Impl) DeleteRepository(_ context.Context, key string) {
	s.RepositoryCache.UpdateEntryRef(key, nil)
}
