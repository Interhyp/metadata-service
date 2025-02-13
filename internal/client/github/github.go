package githubclient

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/Interhyp/metadata-service/internal/util"
	"github.com/google/go-github/v66/github"
	"strconv"
)

type Impl struct {
	CustomConfig config.CustomConfiguration
	client       *github.Client
}

func New(client *github.Client, customConfig config.CustomConfiguration) *Impl {
	return &Impl{
		CustomConfig: customConfig,
		client:       client,
	}
}

func (r *Impl) SetCommitStatusInProgress(ctx context.Context, owner, repoName, commitID, url string, statusKey string) error {
	_, _, err := r.client.Repositories.CreateStatus(ctx, owner, repoName, commitID, &github.RepoStatus{
		State:   util.Ptr("pending"),
		Context: &statusKey,
		URL:     &url,
	})
	return err
}

func (r *Impl) SetCommitStatusSucceeded(ctx context.Context, owner, repoName, commitID, url string, statusKey string) error {
	_, _, err := r.client.Repositories.CreateStatus(ctx, owner, repoName, commitID, &github.RepoStatus{
		State:   util.Ptr("success"),
		Context: &statusKey,
		URL:     &url,
	})
	return err
}

func (r *Impl) SetCommitStatusFailed(ctx context.Context, owner, repoName, commitID, url string, statusKey string) error {
	_, _, err := r.client.Repositories.CreateStatus(ctx, owner, repoName, commitID, &github.RepoStatus{
		State:   util.Ptr("failure"),
		Context: &statusKey,
		URL:     &url,
	})
	return err
}

func (r *Impl) CreatePullRequestComment(ctx context.Context, owner, repoName, pullRequestID, text string) error {
	id, err := strconv.Atoi(pullRequestID)
	if err != nil {
		return err
	}
	_, _, err = r.client.Issues.CreateComment(ctx, owner, repoName, id, &github.IssueComment{
		Body: &text,
	})
	return err
}

func (r *Impl) GetChangedFilesOnPullRequest(ctx context.Context, repoPath, repoName, pullRequestID, toRef string) ([]repository.File, string, error) {
	prId, _ := strconv.Atoi(pullRequestID)
	changes, _, err := r.client.PullRequests.ListFiles(ctx, repoPath, repoName, prId, &github.ListOptions{})
	if err != nil {
		return nil, "", err
	}

	result := make([]repository.File, 0)
	for _, change := range changes {
		contents, _, _, err := r.client.Repositories.GetContents(ctx, repoPath, repoName, change.GetFilename(), &github.RepositoryContentGetOptions{
			Ref: toRef,
		})
		if err != nil {
			// ignore go to next changed file
			continue
		}
		content, _ := contents.GetContent()
		if err != nil || content == "" {
			// ignore go to next changed file
			continue
		}
		result = append(result, repository.File{
			Path:     change.GetFilename(),
			Contents: content,
		})
	}
	return result, toRef, nil
}
