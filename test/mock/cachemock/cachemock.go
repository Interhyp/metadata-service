package ownersmock

import (
	"context"
	openapi "github.com/Interhyp/metadata-service/api"
)

type Mock struct {
}

func (s *Mock) IsCache() bool {
	return true
}

func (s *Mock) SetOwnerListTimestamp(ctx context.Context, timestamp string) {

}

func (s *Mock) GetOwnerListTimestamp(ctx context.Context) string {
	return ""
}

func (s *Mock) GetSortedOwnerAliases(ctx context.Context) []string {
	return nil
}

func (s *Mock) GetOwner(ctx context.Context, alias string) (openapi.OwnerDto, error) {
	if alias == "ownerWithGroup" {
		return openapi.OwnerDto{
			Groups: &map[string][]string{"someGroupName": {"username1", "username2"}},
		}, nil
	}
	return openapi.OwnerDto{}, nil
}

func (s *Mock) PutOwner(ctx context.Context, alias string, entry openapi.OwnerDto) {

}

func (s *Mock) DeleteOwner(ctx context.Context, alias string) {}

func (s *Mock) SetServiceListTimestamp(ctx context.Context, timestamp string) {}

func (s *Mock) GetServiceListTimestamp(ctx context.Context) string {
	return ""
}

func (s *Mock) GetSortedServiceNames(ctx context.Context) []string {
	return nil
}

func (s *Mock) GetService(ctx context.Context, name string) (openapi.ServiceDto, error) {
	return openapi.ServiceDto{}, nil
}

func (s *Mock) PutService(ctx context.Context, name string, entry openapi.ServiceDto) {}

func (s *Mock) DeleteService(ctx context.Context, name string) {

}

func (s *Mock) SetRepositoryListTimestamp(ctx context.Context, timestamp string) {

}

func (s *Mock) GetRepositoryListTimestamp(ctx context.Context) string {
	return ""
}

func (s *Mock) GetSortedRepositoryKeys(ctx context.Context) []string {
	return nil
}

func (s *Mock) GetRepository(ctx context.Context, key string) (openapi.RepositoryDto, error) {
	return openapi.RepositoryDto{}, nil
}

func (s *Mock) PutRepository(ctx context.Context, key string, entry openapi.RepositoryDto) {

}

func (s *Mock) DeleteRepository(ctx context.Context, key string) {

}
