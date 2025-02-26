package githubclient

import (
	"context"
	"fmt"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/google/go-github/v69/github"
)

type Impl struct {
	client *github.Client
}

func New(client *github.Client) *Impl {
	return &Impl{
		client: client,
	}
}

func (r *Impl) StartCheckRun(ctx context.Context, owner, repoName, checkName, sha string) (int64, error) {
	result, _, err := r.client.Checks.CreateCheckRun(ctx, owner, repoName, github.CreateCheckRunOptions{
		Name:    checkName,
		HeadSHA: sha,
		Status:  github.Ptr("in_progress"),
	})
	if err != nil {
		return -1, err
	}
	if result.ID == nil {
		return -1, fmt.Errorf("creating check run '%s' for %s/%s @ %s returned no id", checkName, owner, repoName, sha)
	}
	return result.GetID(), err
}

func (r *Impl) ConcludeCheckRun(ctx context.Context, owner, repoName, checkName string, checkRunId int64, conclusion repository.CheckRunConclusion, details repository.CheckRunDetails) error {
	_, _, err := r.client.Checks.UpdateCheckRun(ctx, owner, repoName, checkRunId, github.UpdateCheckRunOptions{
		Name:       checkName,
		ExternalID: nil,
		Status:     github.Ptr("completed"),
		Conclusion: github.Ptr(string(conclusion)),
		Output: &github.CheckRunOutput{
			Title:   github.Ptr(details.Title),
			Summary: github.Ptr(details.Summary),
			Text:    github.Ptr(details.BodyText),
		},
	})
	if err != nil {
		return err
	}

	return err
}

func (r *Impl) GetChangedFilesForCommit(ctx context.Context, owner, repo, sha string) ([]repository.File, error) {
	commit, _, err := r.client.Repositories.GetCommit(ctx, owner, repo, sha, nil)
	if err != nil {
		return nil, err
	}

	result := make([]repository.File, 0)
	for _, change := range commit.Files {
		contents, _, _, err := r.client.Repositories.GetContents(ctx, owner, repo, change.GetFilename(), &github.RepositoryContentGetOptions{
			Ref: sha,
		})
		if err != nil {
			// ignore go to next changed file
			continue
		}
		content, err := contents.GetContent()
		if err != nil || content == "" {
			// ignore go to next changed file
			continue
		}
		result = append(result, repository.File{
			Path:     change.GetFilename(),
			Contents: content,
		})
	}
	return result, nil
}
