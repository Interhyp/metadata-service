package cache

import (
	"context"
	"github.com/Interhyp/metadata-service/api"
)

const serviceWhat = "service"

func (s *Impl) SetServiceListTimestamp(ctx context.Context, timestamp string) error {
	return s.setTimestamp(ctx, serviceWhat, serviceTimestampKey, timestamp)
}

func (s *Impl) GetServiceListTimestamp(ctx context.Context) (string, error) {
	return s.getTimestamp(ctx, serviceWhat, serviceTimestampKey)
}

func (s *Impl) GetSortedServiceNames(ctx context.Context) ([]string, error) {
	return getSortedKeys(ctx, serviceWhat, s, s.ServiceCache)
}

func (s *Impl) GetService(ctx context.Context, name string) (openapi.ServiceDto, error) {
	return getEntry(ctx, serviceWhat, s, s.ServiceCache, name)
}

func (s *Impl) PutService(ctx context.Context, name string, entry openapi.ServiceDto) error {
	return putEntry(ctx, serviceWhat, s, s.ServiceCache, name, entry)
}

func (s *Impl) DeleteService(ctx context.Context, name string) error {
	return removeEntry(ctx, serviceWhat, s, s.ServiceCache, name)
}
