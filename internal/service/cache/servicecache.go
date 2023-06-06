package cache

import (
	"context"
	"fmt"
	openapi "github.com/Interhyp/metadata-service/api/v1"
	"github.com/StephanHCB/go-backend-service-common/api/apierrors"
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
		return openapi.ServiceDto{}, apierrors.NewNotFoundError("service.notfound", fmt.Sprintf("service %s not found", name), nil, s.Timestamp.Now())
	} else {
		serviceCopy := openapi.ServiceDto{}
		service := (*immutableServicePtr).(openapi.ServiceDto)
		err := deepCopyStruct(service, &serviceCopy)
		return serviceCopy, err
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
