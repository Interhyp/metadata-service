package metadata

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Interhyp/go-backend-service-common/web/middleware/security"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	"io"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	librepo "github.com/Interhyp/go-backend-service-common/acorns/repository"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/Interhyp/metadata-service/internal/acorn/errors/nochangeserror"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Impl struct {
	Configuration       librepo.Configuration
	CustomConfiguration config.CustomConfiguration
	Logging             librepo.Logging
	Timestamp           librepo.Timestamp

	AuthProvider repository.AuthProvider

	GitRepo *git.Repository

	// CommitCacheByFilePath holds information about the newest commit that touches a file, keyed by file path
	CommitCacheByFilePath map[string]repository.CommitInfo

	// KnownCommits remembers which commit hashes we already know
	KnownCommits map[string]bool

	// NewCommits is the list of commits that are new from the most recent Pull operation
	NewCommits []repository.CommitInfo

	// AlreadySeenCommit is the commit hash of the newest commit that we have already cached
	AlreadySeenCommit string

	mu       sync.Mutex
	LastPull time.Time

	consoleOutput bytes.Buffer
}

func New(
	configuration librepo.Configuration,
	customConfig config.CustomConfiguration,
	logging librepo.Logging,
	timestamp librepo.Timestamp,
	authProvider repository.AuthProvider,
) repository.Metadata {
	return &Impl{
		Configuration:       configuration,
		CustomConfiguration: customConfig,
		Logging:             logging,
		Timestamp:           timestamp,
		AuthProvider:        authProvider,

		CommitCacheByFilePath: make(map[string]repository.CommitInfo),
		NewCommits:            make([]repository.CommitInfo, 0),
		KnownCommits:          make(map[string]bool),
	}
}

func (r *Impl) IsMetadata() bool {
	return true
}

func (r *Impl) Setup() error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := r.Clone(ctx); err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to clone service-metadata. BAILING OUT")
		return err
	}

	r.Logging.Logger().Ctx(ctx).Info().Print("successfully set up metadata")
	return nil
}

func (r *Impl) Teardown() {
	ctx := auzerolog.AddLoggerToCtx(context.Background())
	r.Discard(ctx)
}

func (r *Impl) pathsTouchedInCommit(ctx context.Context, commit *object.Commit) ([]string, error) {
	result := make([]string, 0)

	// adapted code from object.StatsContext() because it fails to handle renames and binary files correctly
	fromTree, err := commit.Tree()
	if err != nil {
		return result, err
	}

	toTree := &object.Tree{}
	if commit.NumParents() != 0 {
		firstParent, err := commit.Parents().Next()
		if err != nil {
			return result, err
		}

		toTree, err = firstParent.Tree()
		if err != nil {
			return result, err
		}
	}

	patch, err := toTree.PatchContext(ctx, fromTree)
	if err != nil {
		return result, err
	}

	filePatches := patch.FilePatches()
	for _, filePatch := range filePatches {
		path := ""

		from, to := filePatch.Files()
		if from == nil {
			// New File is created.
			path = to.Path()
		} else if to == nil {
			// File is deleted.
			path = from.Path()
		} else if from.Path() != to.Path() {
			// File is renamed
			path = to.Path()
		} else {
			// Filename unchanged
			path = from.Path()
		}

		if path != "" {
			result = append(result, path)
		}
	}

	return result, nil
}

func (r *Impl) updateCommitCacheMustHoldMutex(ctx context.Context, collectNewCommits bool) error {
	r.Logging.Logger().Ctx(ctx).Info().Printf("rebuilding commit cache")

	if collectNewCommits {
		r.NewCommits = make([]repository.CommitInfo, 0)
	}

	headRef, err := r.GitRepo.Head()
	if err != nil {
		return err
	}

	commitIterator, err := r.GitRepo.Log(&git.LogOptions{
		From:  headRef.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		r.Logging.Logger().Ctx(ctx).Warn().Print("git log failed - console output was: ", r.sanitizedConsoleOutput())
		return err
	}

	r.Logging.Logger().Ctx(ctx).Debug().Print("git log worked - console output was: ", r.sanitizedConsoleOutput())

	seenFileThisRun := make(map[string]bool)

	err = commitIterator.ForEach(func(c *object.Commit) error {
		commitHash := c.Hash.String()
		if commitHash == r.AlreadySeenCommit {
			// stop iteration without raising an error
			return storer.ErrStop
		}

		info := repository.CommitInfo{
			CommitHash: commitHash,
			TimeStamp:  c.Author.When,
			Message:    c.Message,
		}

		r.KnownCommits[commitHash] = true

		pathsTouched, err := r.pathsTouchedInCommit(ctx, c)
		if err != nil {
			return err
		}

		info.FilesChanged = pathsTouched

		for _, path := range pathsTouched {
			_, hasNewer := seenFileThisRun[path]
			if !hasNewer {
				seenFileThisRun[path] = true
				r.CommitCacheByFilePath[path] = info
			}
		}

		if collectNewCommits {
			r.NewCommits = append(r.NewCommits, info)
		}

		return nil
	})

	r.AlreadySeenCommit = headRef.Hash().String()
	return nil
}

func (r *Impl) Clone(ctx context.Context) error {
	r.Logging.Logger().Ctx(ctx).Info().Printf("cloning metadata (git clone)")

	r.mu.Lock()
	defer r.mu.Unlock()

	r.LastPull = r.Timestamp.Now()

	r.CommitCacheByFilePath = make(map[string]repository.CommitInfo)
	r.NewCommits = make([]repository.CommitInfo, 0)
	r.KnownCommits = make(map[string]bool)

	childCtxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	cloneOpts := git.CloneOptions{
		Auth:          r.AuthProvider.ProvideAuth(ctx),
		NoCheckout:    false,
		Progress:      r, // implements io.Writer, sends to Debug logging
		URL:           r.CustomConfiguration.MetadataRepoUrl(),
		ReferenceName: plumbing.ReferenceName(r.CustomConfiguration.MetadataRepoMainline()),
	}

	repo, err := git.CloneContext(childCtxWithTimeout, memory.NewStorage(), memfs.New(), &cloneOpts)
	if err != nil {
		r.Logging.Logger().Ctx(ctx).Warn().Print("git clone failed - console output was: ", r.sanitizedConsoleOutput())
		r.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("git clone failed - returned error: %s", err.Error())

		r.logContextErrorDetails(childCtxWithTimeout, "git clone", "childCtxWithTimeout")
		r.logContextErrorDetails(ctx, "git clone", "ctx")

		return err
	}
	r.GitRepo = repo

	err = r.updateCommitCacheMustHoldMutex(ctx, false)
	if err != nil {
		return err
	}

	r.Logging.Logger().Ctx(ctx).Debug().Print("git clone worked - console output was: ", r.sanitizedConsoleOutput())
	return nil
}

func (r *Impl) Pull(ctx context.Context) error {
	r.Logging.Logger().Ctx(ctx).Info().Printf("updating metadata (git pull)")

	r.mu.Lock()
	defer r.mu.Unlock()

	r.LastPull = r.Timestamp.Now()

	tree, err := r.GitRepo.Worktree()
	if err != nil {
		return err
	}

	childCtxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pullOpts := git.PullOptions{
		Auth:          r.AuthProvider.ProvideAuth(ctx),
		Progress:      r, // implements io.Writer, sends to Debug logging
		RemoteName:    "origin",
		ReferenceName: plumbing.ReferenceName(r.CustomConfiguration.MetadataRepoMainline()),
	}

	err = tree.PullContext(childCtxWithTimeout, &pullOpts)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		r.Logging.Logger().Ctx(ctx).Warn().Print("git pull failed - console output was: ", r.sanitizedConsoleOutput())
		r.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("git pull failed - returned error: %s", err.Error())

		r.logContextErrorDetails(childCtxWithTimeout, "git pull", "childCtxWithTimeout")
		r.logContextErrorDetails(ctx, "git pull", "ctx")

		return err
	}

	err = r.updateCommitCacheMustHoldMutex(ctx, true)
	if err != nil {
		return err
	}

	r.Logging.Logger().Ctx(ctx).Debug().Print("git pull worked - console output was: ", r.sanitizedConsoleOutput())
	return nil
}

func (r *Impl) Commit(ctx context.Context, message string) (repository.CommitInfo, error) {
	r.Logging.Logger().Ctx(ctx).Info().Printf("adding and committing current metadata tree state")

	r.mu.Lock()
	defer r.mu.Unlock()

	commitInfo := repository.CommitInfo{
		CommitHash: "",
		TimeStamp:  r.Timestamp.Now(),
		Message:    message,
	}

	tree, err := r.GitRepo.Worktree()
	if err != nil {
		return commitInfo, err
	}

	err = tree.AddWithOptions(&git.AddOptions{
		All: true,
	})
	if err != nil {
		return commitInfo, err
	}

	// avoid empty commits
	status, err := tree.Status()
	if err != nil {
		return commitInfo, err
	}
	if status.IsClean() {
		return commitInfo, nochangeserror.New(ctx)
	}

	commitTimestamp := r.Timestamp.Now()
	commit, err := tree.Commit(message, &git.CommitOptions{
		Committer: &object.Signature{
			Name:  r.CustomConfiguration.GitCommitterName(),
			Email: r.CustomConfiguration.GitCommitterEmail(),
			When:  commitTimestamp,
		},
		Author: &object.Signature{
			Name:  security.Name(ctx),
			Email: security.Email(ctx),
			When:  commitTimestamp,
		},
	})
	if err != nil {
		return commitInfo, err
	}
	commitInfo.CommitHash = commit.String()
	commitInfo.TimeStamp = commitTimestamp

	err = r.updateCommitCacheMustHoldMutex(ctx, false)
	if err != nil {
		return commitInfo, err
	}

	return commitInfo, nil
}

func (r *Impl) Push(ctx context.Context) error {
	r.Logging.Logger().Ctx(ctx).Info().Printf("pushing metadata to upstream (git push)")

	r.mu.Lock()
	defer r.mu.Unlock()

	childCtxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pushOpts := git.PushOptions{
		Auth:       r.AuthProvider.ProvideAuth(ctx),
		Progress:   r, // implements io.Writer, sends to Debug logging
		RemoteName: "origin",
	}

	err := r.GitRepo.PushContext(childCtxWithTimeout, &pushOpts)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		r.Logging.Logger().Ctx(ctx).Warn().Print("git push failed - console output was: ", r.sanitizedConsoleOutput())
		r.Logging.Logger().Ctx(ctx).Warn().WithErr(err).Printf("git push failed - returned error: %s", err.Error())

		r.logContextErrorDetails(childCtxWithTimeout, "git push", "childCtxWithTimeout")
		r.logContextErrorDetails(ctx, "git push", "ctx")

		return err
	}

	r.LastPull = r.Timestamp.Now()

	r.Logging.Logger().Ctx(ctx).Debug().Print("git push worked - console output was: ", r.sanitizedConsoleOutput())
	return nil
}

func (r *Impl) Discard(ctx context.Context) {
	r.Logging.Logger().Ctx(ctx).Info().Printf("discarding metadata clone")

	r.mu.Lock()
	defer r.mu.Unlock()

	r.GitRepo = nil

	r.CommitCacheByFilePath = make(map[string]repository.CommitInfo)
	r.NewCommits = make([]repository.CommitInfo, 0)
	r.KnownCommits = make(map[string]bool)
}

func (r *Impl) logContextErrorDetails(ctx context.Context, operation string, contextName string) {
	ctxCause := context.Cause(ctx)
	if ctxCause != nil {
		r.Logging.Logger().Ctx(ctx).Warn().WithErr(ctxCause).Printf("%s failed - %s cancellation cause: %s", operation, contextName, ctxCause.Error())
	} else {
		r.Logging.Logger().Ctx(ctx).Warn().Printf("%s failed - %s not cancelled", operation, contextName)
	}
}

func (r *Impl) LastUpdated() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.LastPull
}

func (r *Impl) NewPulledCommits() []repository.CommitInfo {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]repository.CommitInfo, len(r.NewCommits))
	_ = copy(result, r.NewCommits)

	return result
}

func (r *Impl) IsCommitKnown(hash string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.KnownCommits[hash]
	return ok
}

func (r *Impl) Stat(filename string) (os.FileInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	tree, err := r.GitRepo.Worktree()
	if err != nil {
		return nil, err
	}

	return tree.Filesystem.Stat(filename)
}

func (r *Impl) ReadDir(path string) ([]os.FileInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	tree, err := r.GitRepo.Worktree()
	if err != nil {
		return nil, err
	}

	return tree.Filesystem.ReadDir(path)
}

func (r *Impl) ReadFile(filename string) ([]byte, repository.CommitInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	errorCommitInfo := repository.CommitInfo{
		CommitHash: "",
		TimeStamp:  r.Timestamp.Now(),
		Message:    "",
	}

	tree, err := r.GitRepo.Worktree()
	if err != nil {
		return nil, errorCommitInfo, err
	}

	fileHandle, err := tree.Filesystem.Open(filename)
	if err != nil {
		return nil, errorCommitInfo, err
	}
	defer fileHandle.Close()

	data, err := io.ReadAll(fileHandle)
	if err != nil {
		return nil, errorCommitInfo, err
	}

	commitInfo, ok := r.CommitCacheByFilePath[filename]
	if !ok {
		return nil, errorCommitInfo, fmt.Errorf("failed to find commit info on %s", filename)
	}

	return data, commitInfo, nil
}

func (r *Impl) WriteFile(filename string, contents []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tree, err := r.GitRepo.Worktree()
	if err != nil {
		return err
	}

	fileHandle, err := tree.Filesystem.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fileHandle.Close()

	_, err = fileHandle.Write(contents)
	if err != nil {
		return err
	}

	return nil
}

func (r *Impl) DeleteFile(filename string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tree, err := r.GitRepo.Worktree()
	if err != nil {
		return err
	}

	err = tree.Filesystem.Remove(filename)
	if err != nil {
		return err
	}

	// add with all: true does not add deletions (during commit), so git add here

	_, err = tree.Remove(filename)
	if err != nil {
		return err
	}

	return nil
}

func (r *Impl) MkdirAll(path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tree, err := r.GitRepo.Worktree()
	if err != nil {
		return err
	}

	err = tree.Filesystem.MkdirAll(path, 0755)
	return err
}

// implement io.Writer so r can be used by git for logging

func (r *Impl) Write(p []byte) (n int, err error) {
	r.consoleOutput.Write(p)
	return n, nil
}

func (r *Impl) sanitizedConsoleOutput() string {
	return strings.Map(func(c rune) rune {
		if unicode.IsGraphic(c) {
			return c
		} else {
			return -1 // drop
		}
	}, r.consoleOutput.String())
}
