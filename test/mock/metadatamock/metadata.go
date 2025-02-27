package metadatamock

import (
	"context"
	"errors"
	"fmt"
	"github.com/Interhyp/metadata-service/test/mock/checkoutmock"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Interhyp/metadata-service/internal/acorn/errors/nochangeserror"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"

	"github.com/Interhyp/go-backend-service-common/api/apierrors"
	"github.com/go-git/go-billy/v5"
)
import _ "github.com/go-git/go-git/v5"

type Impl struct {
	Fs  billy.Filesystem
	Now func() time.Time

	FilesWritten   map[string]bool
	FilesCommitted map[string]bool
	Pushed         bool
	InvalidIssue   bool

	SimulateRemoteFailure      bool
	SimulateConcurrencyFailure bool
	SimulateUnchangedFailure   bool
}

func New() repository.Metadata {
	return &Impl{
		Now: time.Now,
	}
}

func (r *Impl) Setup() error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	if err := r.Clone(ctx); err != nil {
		return err
	}
	return nil
}

func (r *Impl) Teardown() {
}

func (r *Impl) IsMetadata() bool {
	return true
}

const (
	origCommitHash    = "6c8ac2c35791edf9979623c717a243fc53400000"
	newCommitHash     = "6c8ac2c35791edf9979623c717a2430000000000"
	origCommitMessage = "ISSUE-0000: original"
	newCommitMessage  = "ISSUE-2345: new"
)

func (r *Impl) commitHash(filename string) string {
	_, found := r.FilesCommitted[filename]
	if found {
		return newCommitHash
	} else {
		return origCommitHash
	}
}

func (r *Impl) commitMessage(filename string) string {
	_, found := r.FilesCommitted[filename]
	if found {
		return newCommitMessage
	} else {
		return origCommitMessage
	}
}

func (r *Impl) Clone(ctx context.Context) error {
	fs, err := checkoutmock.New()
	if err != nil {
		return err
	}
	r.Fs = fs
	r.FilesCommitted = make(map[string]bool)
	r.FilesWritten = make(map[string]bool)
	r.SimulateRemoteFailure = false
	r.SimulateConcurrencyFailure = false
	r.SimulateUnchangedFailure = false
	r.Pushed = false
	r.InvalidIssue = false
	return nil
}

func (r *Impl) Pull(ctx context.Context) error {
	return nil
}

func (r *Impl) Commit(ctx context.Context, message string) (repository.CommitInfo, error) {
	commitInfo := repository.CommitInfo{
		CommitHash: "",
		TimeStamp:  r.Now(),
		Message:    "",
	}

	if r.SimulateUnchangedFailure {
		return commitInfo, nochangeserror.New(ctx)
	}
	if strings.Contains(message, "INVALID-12345") {
		r.InvalidIssue = true
	}

	r.FilesCommitted = r.FilesWritten
	commitInfo.CommitHash = newCommitHash
	commitInfo.Message = message
	return commitInfo, nil
}

func (r *Impl) Push(ctx context.Context) error {
	if r.SimulateRemoteFailure {
		return apierrors.NewBadGatewayError("downstream.unavailable", "the git server is currently unavailable or failed to service the request", nil, r.Now())
	}
	if r.SimulateConcurrencyFailure {
		return apierrors.NewConflictError("", "cannot push", nil, r.Now())
	}
	if r.InvalidIssue {
		return fmt.Errorf("failed to push ref: pre-receive hook declined: something something")
	}
	r.Pushed = true
	return nil
}

func (r *Impl) Discard(ctx context.Context) {
}

func (r *Impl) LastUpdated() time.Time {
	return r.Now()
}

func (r *Impl) NewPulledCommits() []repository.CommitInfo {
	return make([]repository.CommitInfo, 0)
}

func (r *Impl) IsCommitKnown(hash string) bool {
	return false
}

func (r *Impl) Stat(filename string) (os.FileInfo, error) {
	return r.Fs.Stat(filename)
}

func (r *Impl) ReadDir(path string) ([]os.FileInfo, error) {
	return r.Fs.ReadDir(path)
}

func (r *Impl) ReadFile(filename string) ([]byte, repository.CommitInfo, error) {
	commitInfo := repository.CommitInfo{
		CommitHash: "",
		TimeStamp:  r.Now(),
		Message:    "",
	}

	fileHandle, err := r.Fs.Open(filename)
	if err != nil {
		return nil, commitInfo, err
	}
	defer fileHandle.Close()

	data, err := io.ReadAll(fileHandle)
	if err != nil {
		return nil, commitInfo, err
	}

	commitInfo.CommitHash = r.commitHash(filename)
	commitInfo.Message = r.commitMessage(filename)

	return data, commitInfo, nil
}

func (r *Impl) WriteFile(filename string, contents []byte) error {
	fileHandle, err := r.Fs.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fileHandle.Close()

	_, err = fileHandle.Write(contents)
	if err != nil {
		return err
	}

	r.FilesWritten[filename] = true
	return nil
}

func (r *Impl) DeleteFile(filename string) error {
	err := r.Fs.Remove(filename)
	if err != nil {
		return err
	}
	r.FilesWritten[filename] = true
	return nil
}

func (r *Impl) MkdirAll(path string) error {
	return r.Fs.MkdirAll(path, 0755)
}

// reset for the next test

func (r *Impl) Reset() {
	_ = r.Clone(context.TODO())
}

func (r *Impl) ReadContents(filename string) string {
	by, _, err := r.ReadFile(filename)
	if errors.Is(os.ErrNotExist, err) {
		return "<notfound>"
	}
	if by == nil {
		return "<nil>"
	}
	return string(by)
}
