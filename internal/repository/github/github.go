package githubclient

import (
	"context"
	"errors"
	"fmt"
	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/google/go-github/v69/github"
)

type Impl struct {
	client    *github.Client
	Timestamp librepo.Timestamp
}

func New(timestamp librepo.Timestamp, client *github.Client) *Impl {
	return &Impl{
		Timestamp: timestamp,
		client:    client,
	}
}

func (r *Impl) StartCheckRun(ctx context.Context, owner, repoName, checkName, sha string) (int64, error) {
	result, _, err := r.client.Checks.CreateCheckRun(ctx, owner, repoName, github.CreateCheckRunOptions{
		Name:    checkName,
		HeadSHA: sha,
		StartedAt: &github.Timestamp{
			Time: r.Timestamp.Now(),
		},
		Status: github.Ptr("in_progress"),
	})
	if err != nil {
		return -1, err
	}
	if result.ID == nil {
		return -1, fmt.Errorf("creating check run '%s' for %s/%s @ %s returned no id", checkName, owner, repoName, sha)
	}
	return result.GetID(), err
}

func (r *Impl) ConcludeCheckRun(ctx context.Context, owner, repoName, checkName string, checkRunId int64, conclusion repository.CheckRunConclusion, output github.CheckRunOutput) error {
	annotationLimit := 50
	annotations := output.Annotations
	errs := make([]error, 0)
	for len(annotations) > annotationLimit {
		batch := annotations[:annotationLimit]
		annotations = annotations[annotationLimit:]
		_, _, err := r.client.Checks.UpdateCheckRun(ctx, owner, repoName, checkRunId, github.UpdateCheckRunOptions{
			Name:   checkName,
			Status: github.Ptr("in_progress"),
			Output: &github.CheckRunOutput{
				Title:       output.Title,
				Summary:     output.Summary,
				Annotations: batch,
			},
		})
		errs = append(errs, err)
	}
	text := output.Text
	if text != nil {
		runes := []rune(*text)
		// If body is longer than 65535 chars, Github returns 422 Invalid request with message "Only 65535 characters are allowed; 79127 were supplied."
		ghCharLimit := 65535
		if len(runes) > ghCharLimit {
			warning := []rune("# :warning: Too many errors for one message. Fix issues below and run check again against fixed commit.\n")
			maxLength := ghCharLimit - len(warning)
			updated := string(append(warning, runes[:maxLength]...))
			text = &updated
		}
	}
	_, _, err := r.client.Checks.UpdateCheckRun(ctx, owner, repoName, checkRunId, github.UpdateCheckRunOptions{
		Name:       checkName,
		Status:     github.Ptr("completed"),
		Conclusion: github.Ptr(string(conclusion)),
		CompletedAt: &github.Timestamp{
			Time: r.Timestamp.Now(),
		},
		Output: &github.CheckRunOutput{
			Title:       output.Title,
			Summary:     output.Summary,
			Text:        text,
			Annotations: annotations,
			Images:      output.Images,
		},
	})
	errs = append(errs, err)
	return errors.Join(errs...)
}
