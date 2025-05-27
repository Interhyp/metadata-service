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
	FixAction     = "fix-all"
	ActionTimeout = 1 * time.Minute
)

func (h *Impl) PerformRequestedAction(ctx context.Context, requestedAction string, checkRun *github.CheckRun, requestingUser *github.User) error {
	aulogging.Logger.Ctx(ctx).Info().Printf("received webhook for requested_action %s (suite: %d|run: %d)", requestedAction, checkRun.CheckSuite.GetID(), checkRun.GetID())
	independentCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), ActionTimeout)
	defer cancel()

	switch requestedAction {
	case FixAction:
		msg := "formatting files/adding missing exemptions"
		fixFunc := func(branchName string, worktree *git.Worktree) error {
			aulogging.Logger.Ctx(ctx).Debug().Printf("%s on branch %s", msg, branchName)
			err := MetadataYamlFileWalker(worktree.Filesystem,
				WithIndentation(h.CustomConfiguration.YamlIndentation()),
			).FormatMetadata()
			if err != nil {
				return err
			}
			err = MetadataYamlFileWalker(worktree.Filesystem,
				WithExpectedExemptions(h.CustomConfiguration.CheckedExpectedExemptions()),
			).FixExemptions()
			return err
		}
		return h.commitFixes(independentCtx, fixFunc, checkRun.GetCheckSuite().GetHeadBranch(), requestingUser.GetLogin(), msg)
	}

	aulogging.Logger.Ctx(independentCtx).Info().Printf("successfully processed webhook for requested_action %s (suite: %d|run: %d)", requestedAction, checkRun.CheckSuite.GetID(), checkRun.GetID())
	return nil
}

func (h *Impl) commitFixes(ctx context.Context, fixFunc func(branchName string, worktree *git.Worktree) error, branchName string, user string, msg string) error {
	if branchName == "" {
		return fmt.Errorf("missing branch name for fixing %s", msg)
	}
	aulogging.Logger.Ctx(ctx).Info().Printf("start fixing '%s' on branch %s", msg, branchName)
	author, err := h.Github.GetUser(ctx, user)
	if err != nil {
		return err
	}

	aulogging.Logger.Ctx(ctx).Debug().Printf("cloning branch %s", branchName)
	repo, worktree, err := h.clone(ctx, branchName)
	if err != nil {
		return err
	}

	err = fixFunc(branchName, worktree)
	if err != nil {
		return err
	}

	aulogging.Logger.Ctx(ctx).Debug().Printf("committing %s onto branch %s", msg, branchName)
	err = h.commit(worktree, author, msg)
	if err != nil {
		return err
	}

	aulogging.Logger.Ctx(ctx).Debug().Printf("pushing %s onto branch %s", msg, branchName)
	err = repo.PushContext(ctx, &git.PushOptions{
		Auth:       h.AuthProvider.ProvideAuth(ctx),
		RemoteName: "origin",
	})
	if err != nil {
		return err
	}
	aulogging.Logger.Ctx(ctx).Info().Printf("finished fixing '%s' on branch %s", msg, branchName)
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

func (h *Impl) commit(worktree *git.Worktree, author *github.User, msg string) error {
	commitTimestamp := h.timestamp.Now()
	commitMsg := fmt.Sprintf("%s%s", h.CustomConfiguration.FormattingActionCommitMsgPrefix(), msg)
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
