package bitbucket

import (
	"context"
	"fmt"
	"github.com/Interhyp/metadata-service/internal/acorn/errors/httperror"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/repository/bitbucket/bbclient"
	"github.com/Interhyp/metadata-service/internal/repository/bitbucket/bbclientint"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"net/http"
	"sort"
)

type Impl struct {
	Configuration librepo.Configuration
	Logging       librepo.Logging
	Vault         librepo.Vault

	LowLevel bbclientint.BitbucketClient
}

func New(
	configuration librepo.Configuration,
	logging librepo.Logging,
	vault librepo.Vault,
) repository.Bitbucket {
	return &Impl{
		Configuration: configuration,
		Logging:       logging,
		Vault:         vault,
		LowLevel:      bbclient.New(configuration, logging, vault),
	}
}

func (r *Impl) IsBitbucket() bool {
	return true
}

func (r *Impl) Setup() error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := r.SetupClient(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to set up bitbucket client. BAILING OUT")
		return err
	}

	r.Logging.Logger().Ctx(ctx).Info().Print("successfully set up bitbucket")
	return nil
}

func (r *Impl) SetupClient(ctx context.Context) error {
	r.Logging.Logger().Ctx(ctx).Info().Print("setting up bitbucket client")
	return r.LowLevel.Setup()
}

func (r *Impl) GetBitbucketUser(ctx context.Context, username string) (repository.BitbucketUser, error) {
	response, err := r.LowLevel.GetBitbucketUser(ctx, username)
	if err != nil {
		return repository.BitbucketUser{}, err
	}
	return response, nil
}

func (r *Impl) GetBitbucketUsers(ctx context.Context, usernames []string) ([]repository.BitbucketUser, error) {
	users := make([]repository.BitbucketUser, 0)
	for _, username := range usernames {
		bbUser, err := r.GetBitbucketUser(ctx, username)
		if err != nil {
			if httperror.Is(err) && err.(*httperror.Impl).Status() == http.StatusNotFound {
				r.Logging.Logger().Ctx(ctx).Warn().Printf("bitbucket user %s does not exist", username)
				continue
			}
			r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print(fmt.Sprintf("failed to read bitbucket user %s", username))
			return make([]repository.BitbucketUser, 0), err
		}
		users = append(users, bbUser)
	}
	return users, nil
}

func (r *Impl) FilterExistingUsernames(ctx context.Context, usernames []string) ([]string, error) {
	existingUsernames := make([]string, 0)

	dedupUsers := Unique(usernames)
	existingUsers, err := r.GetBitbucketUsers(ctx, dedupUsers)
	if err != nil {
		return existingUsernames, err
	}

	for _, user := range existingUsers {
		existingUsernames = append(existingUsernames, user.Name)
	}
	sort.Strings(existingUsernames)

	return existingUsernames, nil
}

func Unique[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := make([]T, 0)
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
