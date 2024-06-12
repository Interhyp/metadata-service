package bitbucket

import (
	"context"
	"fmt"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/errors/httperror"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/repository/bitbucket/bbclient"
	"github.com/Interhyp/metadata-service/internal/repository/bitbucket/bbclientint"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"net/http"
	"sort"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Vault               librepo.Vault

	LowLevel bbclientint.BitbucketClient
}

func New(
	configuration librepo.Configuration,
	customConfiguration config.CustomConfiguration,
	logging librepo.Logging,
	vault librepo.Vault,
) repository.Bitbucket {
	return &Impl{
		Configuration:       configuration,
		CustomConfiguration: customConfiguration,
		Logging:             logging,
		Vault:               vault,
		LowLevel:            bbclient.New(configuration, logging, vault),
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

func (r *Impl) GetChangedFilesOnPullRequest(ctx context.Context, pullRequestId int) ([]repository.File, string, error) {
	aulogging.Logger.Ctx(ctx).Info().Printf("obtaining changes for pull request %d", pullRequestId)

	project := r.CustomConfiguration.MetadataRepoProject()
	slug := r.CustomConfiguration.MetadataRepoName()
	pullRequest, err := r.LowLevel.GetPullRequest(ctx, project, slug, int32(pullRequestId))
	if err != nil {
		return nil, "", err
	}

	prSourceHead := pullRequest.FromRef.LatestCommit
	changes, err := r.LowLevel.GetChanges(ctx, project, slug, pullRequest.ToRef.LatestCommit, prSourceHead)
	if err != nil {
		return nil, prSourceHead, err
	}

	aulogging.Logger.Ctx(ctx).Info().Printf("pull request had %d changed files", len(changes.Values))

	result := make([]repository.File, 0)
	for _, change := range changes.Values {
		contents, err := r.LowLevel.GetFileContentsAt(ctx, project, slug, prSourceHead, change.Path.ToString)
		if err != nil {
			asHttpError, ok := err.(httperror.Error)
			if ok && asHttpError.Status() == http.StatusNotFound {
				aulogging.Logger.Ctx(ctx).Debug().Printf("path %s not present on PR head - skipping and continuing", change.Path.ToString)
				continue // expected situation - happens for deleted files, or for files added on mainline after fork (which show up in changes)
			} else {
				aulogging.Logger.Ctx(ctx).Info().Printf("failed to retrieve change for %s: %s", change.Path.ToString, err.Error())
				return nil, prSourceHead, err
			}
		}

		result = append(result, repository.File{
			Path:     change.Path.ToString,
			Contents: contents,
		})
	}

	aulogging.Logger.Ctx(ctx).Info().Printf("successfully obtained %d changes for pull request %d", len(result), pullRequestId)
	return result, prSourceHead, nil
}

func (r *Impl) AddCommitBuildStatus(ctx context.Context, commitHash string, url string, key string, success bool) error {
	project := r.CustomConfiguration.MetadataRepoProject()
	slug := r.CustomConfiguration.MetadataRepoName()

	state := "FAILED"
	if success {
		state = "SUCCESS"
	}

	request := bbclientint.CommitBuildStatusRequest{
		Key:   key,
		State: state,
		Url:   url,
	}

	return r.LowLevel.AddProjectRepositoryCommitBuildStatus(ctx, project, slug, commitHash, request)
}

func (r *Impl) CreatePullRequestComment(ctx context.Context, pullRequestId int, comment string) error {
	project := r.CustomConfiguration.MetadataRepoProject()
	slug := r.CustomConfiguration.MetadataRepoName()

	request := bbclientint.PullRequestCommentRequest{
		Text: comment,
	}

	_, err := r.LowLevel.CreatePullRequestComment(ctx, project, slug, int64(pullRequestId), request)
	return err
}
