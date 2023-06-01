package cache

import (
	"context"
	"fmt"
	openapi "github.com/Interhyp/metadata-service/api/v1"
	"github.com/StephanHCB/go-backend-service-common/api/apierrors"
)

func (s *Impl) SetOwnerListTimestamp(_ context.Context, timestamp string) {
	s.OwnerCache.SetTimestamp(timestamp)
}

func (s *Impl) GetOwnerListTimestamp(_ context.Context) string {
	return s.OwnerCache.GetTimestamp()
}

func (s *Impl) GetSortedOwnerAliases(_ context.Context) []string {
	keysPtr := s.OwnerCache.GetSortedKeys()
	return deepCopyStringSlice(*keysPtr)
}

func (s *Impl) GetOwner(ctx context.Context, alias string) (openapi.OwnerDto, error) {
	immutableOwnerPtr := s.OwnerCache.GetEntryRef(alias)
	if immutableOwnerPtr == nil {
		s.Logging.Logger().Ctx(ctx).Info().Printf("owner %v not found", alias)
		return openapi.OwnerDto{}, apierrors.NewNotFoundError("owner.notfound", fmt.Sprintf("owner %s not found", alias), nil, s.Timestamp.Now())
	} else {
		ownerCopy := openapi.OwnerDto{}
		owner := (*immutableOwnerPtr).(openapi.OwnerDto)
		err := deepCopyStruct(owner, &ownerCopy)
		return ownerCopy, err
	}
}

func (s *Impl) PutOwner(_ context.Context, alias string, entry openapi.OwnerDto) {
	var e interface{}
	e = entry

	s.OwnerCache.UpdateEntryRef(alias, &e)
}

func (s *Impl) DeleteOwner(_ context.Context, alias string) {
	s.OwnerCache.UpdateEntryRef(alias, nil)
	// TODO since this may come in from reading a manually made git commit, in this lowlevel cache we cascade
	//
	// s.scDeleteOwner(alias)
	// s.rcDeleteOwner(alias)
}
