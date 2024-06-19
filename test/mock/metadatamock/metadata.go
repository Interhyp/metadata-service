package metadatamock

import (
	"context"
	"errors"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	"io"
	"os"
	"time"

	"github.com/Interhyp/metadata-service/internal/acorn/errors/nochangeserror"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"

	"github.com/StephanHCB/go-backend-service-common/api/apierrors"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
)
import _ "github.com/go-git/go-git/v5"

type Impl struct {
	Fs  billy.Filesystem
	Now func() time.Time

	FilesWritten   map[string]bool
	FilesCommitted map[string]bool
	Pushed         bool

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

const ownerInfo = `contact: somebody@some-organisation.com
teamsChannelURL: https://teams.microsoft.com/l/channel/somechannel
productOwner: kschlangenheldt
defaultJiraProject: ISSUE
`

const ownerInfoNoPromoters = `contact: somebody@some-organisation.com
teamsChannelURL: https://teams.microsoft.com/l/channel/somechannel
productOwner: kschlangenheldt
defaultJiraProject: ISSUE
`

const service = `quicklinks:
- title: Swagger UI
  url: /swagger-ui/index.html
repositories:
- some-service-backend/helm-deployment
- some-service-backend/implementation
alertTarget: https://webhook.com/9asdflk29d4m39g
developmentOnly: false
`

const deployment = `mainline: main
url: ssh://git@bitbucket.some-organisation.com:7999/PROJECT/some-service-backend-deployment.git
deployment:
  kubernetes:
    instances:
    - namespace: project
      environment: prod
      cluster: openshift
    - namespace: project
      environment: dev
      cluster: openshift
    - namespace: project
      environment: test
      cluster: openshift
    - namespace: project
      environment: livetest
      cluster: openshift
generator: third-party-software
configuration:
  accessKeys:
  - key: DEPLOYMENT
    permission: REPO_READ
  - data: 'ssh-key abcdefgh.....'
    permission: REPO_WRITE
  commitMessageType: DEFAULT
  mergeConfig:
    defaultStrategy:
      id: "no-ff"
    strategies:
      - id: "no-ff"
      - id: "ff"
      - id: "ff-only"
      - id: "squash"
  requireIssue: true
  approvers:
    testing:
    - some-user
`

const deployment2 = `mainline: main
url: ssh://git@bitbucket.some-organisation.com:7999/PROJECT/whatever-deployment.git
generator: third-party-software
filecategory:
  forbidden-key:
    - some/interesting/file.txt
`

const implementation = `mainline: master
url: ssh://git@bitbucket.some-organisation.com:7999/PROJECT/some-service-backend.git
generator: java-spring-cloud
`

const implementation2 = `mainline: master
url: ssh://git@bitbucket.some-organisation.com:7999/PROJECT/whatever.git
generator: java-spring-cloud
`

const chart = `url: ssh://git@bitbucket.some-organisation.com:7999/helm/karma-wrapper.git
mainline: master
unittest: false
`

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

func (r *Impl) writeFile(filename string, contents string) error {
	f, err := r.Fs.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(contents))
	return err
}

func (r *Impl) Clone(ctx context.Context) error {
	r.Fs = memfs.New()
	err := r.Fs.MkdirAll("owners/some-owner/services", 0755)
	if err != nil {
		return err
	}
	err = r.Fs.MkdirAll("owners/some-owner/repositories", 0755)
	if err != nil {
		return err
	}
	err = r.Fs.MkdirAll("owners/deleteme/services", 0755)
	if err != nil {
		return err
	}
	err = r.Fs.MkdirAll("owners/deleteme/repositories", 0755)
	if err != nil {
		return err
	}

	err = r.writeFile("owners/some-owner/owner.info.yaml", ownerInfo)
	if err != nil {
		return err
	}
	err = r.writeFile("owners/some-owner/services/some-service-backend.yaml", service)
	if err != nil {
		return err
	}
	err = r.writeFile("owners/some-owner/repositories/some-service-backend.helm-deployment.yaml", deployment)
	if err != nil {
		return err
	}
	err = r.writeFile("owners/some-owner/repositories/some-service-backend.implementation.yaml", implementation)
	if err != nil {
		return err
	}
	err = r.writeFile("owners/some-owner/repositories/whatever.helm-deployment.yaml", deployment2)
	if err != nil {
		return err
	}
	err = r.writeFile("owners/some-owner/repositories/whatever.implementation.yaml", implementation2)
	if err != nil {
		return err
	}
	err = r.writeFile("owners/some-owner/repositories/karma-wrapper.helm-chart.yaml", chart)
	if err != nil {
		return err
	}

	err = r.writeFile("owners/deleteme/owner.info.yaml", ownerInfoNoPromoters)
	if err != nil {
		return err
	}

	r.FilesCommitted = make(map[string]bool)
	r.FilesWritten = make(map[string]bool)
	r.SimulateRemoteFailure = false
	r.SimulateConcurrencyFailure = false
	r.SimulateUnchangedFailure = false
	r.Pushed = false
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
