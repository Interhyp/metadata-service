package cache

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/errors/nosuchserviceerror"
	openapi "github.com/Interhyp/metadata-service/api/v1"
)

func (s *Impl) SetServiceListTimestamp(_ context.Context, timestamp string) {
	s.ServiceCache.SetTimestamp(timestamp)
}

func (s *Impl) GetServiceListTimestamp(_ context.Context) string {
	return s.ServiceCache.GetTimestamp()
}

func (s *Impl) GetSortedServiceNames(_ context.Context) []string {
	keysPtr := s.ServiceCache.GetSortedKeys()
	return deepCopyStringSlice(*keysPtr)
}

func (s *Impl) GetService(ctx context.Context, name string) (openapi.ServiceDto, error) {
	immutableServicePtr := s.ServiceCache.GetEntryRef(name)
	if immutableServicePtr == nil {
		return openapi.ServiceDto{}, nosuchserviceerror.New(ctx, name)
	} else {
		return deepCopyService(immutableServicePtr), nil
	}
}

func (s *Impl) PutService(_ context.Context, name string, entry openapi.ServiceDto) {
	var e interface{}
	e = entry

	s.ServiceCache.UpdateEntryRef(name, &e)
}

func (s *Impl) DeleteService(_ context.Context, name string) {
	s.ServiceCache.UpdateEntryRef(name, nil)
}
