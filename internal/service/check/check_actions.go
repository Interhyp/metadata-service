package check

import (
	"context"
	"fmt"
	"github.com/StephanHCB/go-autumn-logging"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-github/v70/github"
	"time"
)

const (
	FixFormattingAction = "fix-formatting"
	ActionTimeout       = 1 * time.Minute
)

func (h *Impl) PerformRequestedAction(ctx context.Context, requestedAction string, checkRun *github.CheckRun, requestingUser *github.User) error {
	aulogging.Logger.Ctx(ctx).Info().Printf("received webhook for requested_action %s (suite: %d|run: %d)", requestedAction, checkRun.CheckSuite.GetID(), checkRun.GetID())
	independentCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), ActionTimeout)
	defer cancel()

	switch requestedAction {
	case FixFormattingAction:
		return h.commitFormatFixes(independentCtx, checkRun.GetCheckSuite().GetHeadBranch(), requestingUser.GetLogin())
	}

	aulogging.Logger.Ctx(independentCtx).Info().Printf("successfully processed webhook for requested_action %s (suite: %d|run: %d)", requestedAction, checkRun.CheckSuite.GetID(), checkRun.GetID())
	return nil
}

func (h *Impl) commitFormatFixes(ctx context.Context, branchName string, user string) error {
	if branchName == "" {
		return fmt.Errorf("missing branch name for formatting files")
	}
	aulogging.Logger.Ctx(ctx).Info().Printf("start fixing file format on branch %s", branchName)
	author, err := h.Github.GetUser(ctx, user)
	if err != nil {
		return err
	}

	aulogging.Logger.Ctx(ctx).Debug().Printf("cloning branch %s", branchName)
	repo, worktree, err := h.clone(ctx, branchName)
	if err != nil {
		return err
	}

	aulogging.Logger.Ctx(ctx).Debug().Printf("formatting files on branch %s", branchName)
	err = MetadataYamlFileWalker(worktree.Filesystem, h.CustomConfiguration.YamlIndentation()).FormatMetadata()
	if err != nil {
		return err
	}

	aulogging.Logger.Ctx(ctx).Debug().Printf("committing formatted files onto branch %s", branchName)
	err = h.commit(worktree, author)
	if err != nil {
		return err
	}

	aulogging.Logger.Ctx(ctx).Debug().Printf("pushing formatted files onto branch %s", branchName)
	err = repo.PushContext(ctx, &git.PushOptions{
		Auth:       h.AuthProvider.ProvideAuth(ctx),
		RemoteName: "origin",
	})
	if err != nil {
		return err
	}
	aulogging.Logger.Ctx(ctx).Info().Printf("finished fixing file format on branch %s", branchName)
	return nil
}

func (h *Impl) clone(ctx context.Context, branchName string) (*git.Repository, *git.Worktree, error) {
	branch := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName))
	repo, err := git.CloneContext(ctx, memory.NewStorage(), memfs.New(), &git.CloneOptions{
		Auth:          h.AuthProvider.ProvideAuth(ctx),
		NoCheckout:    false,
		URL:           h.CustomConfiguration.MetadataRepoUrl(),
		ReferenceName: branch,
	})
	if err != nil {
		return nil, nil, err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, nil, err
	}
	return repo, worktree, nil
}

func (h *Impl) commit(worktree *git.Worktree, author *github.User) error {
	commitTimestamp := h.timestamp.Now()
	commitMsg := fmt.Sprintf("%sFormat files", h.CustomConfiguration.FormattingActionCommitMsgPrefix())
	_, err := worktree.Commit(commitMsg, &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name:  author.GetName(),
			Email: author.GetEmail(),
			When:  commitTimestamp,
		},
		Committer: &object.Signature{
			Name:  h.CustomConfiguration.GitCommitterName(),
			Email: h.CustomConfiguration.GitCommitterEmail(),
			When:  commitTimestamp,
		},
	})
	return err
}
