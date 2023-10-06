package repository

import (
	"context"
	"os"
	"time"
)

// CommitInfo holds information about a commit.
type CommitInfo struct {
	CommitHash   string
	TimeStamp    time.Time
	Message      string
	FilesChanged []string
}

// Metadata is the central singleton representing the service-metadata git repository.
//
// All operations are protected by a mutex, but of course this does not prevent multiple
// goroutines from making changes between operations, so you will probably need a higher level
// mutex to avoid inadvertently committing changes made by another goroutine.
type Metadata interface {
	IsMetadata() bool

	Setup() error
	Teardown()

	// Clone performs an initial in-memory clone of the metadata repository on the mainline
	Clone(ctx context.Context) error

	// Pull updates the in-memory clone of the metadata repository on the mainline
	//
	// Any new commits that were not previously seen can now be obtained by NewPulledCommits.
	Pull(ctx context.Context) error

	// Commit performs a local add all and commit and returns the commit hash and the timestamp
	//
	// note: if this fails, the repository may be in an inconsistent state, so you should
	// Discard and Clone it again.
	Commit(ctx context.Context, message string) (CommitInfo, error)

	// Push sends commits from the in-memory clone to the upstream
	Push(ctx context.Context) error

	// Discard the in-memory clone (cannot fail, but will leave memory allocated until garbage collection)
	//
	// note: doing a new Clone implicitly discards
	Discard(ctx context.Context)

	// LastUpdated gives the time the git repo was last pulled (or pushed, which also ensures it is up-to-date).
	LastUpdated() time.Time

	// NewPulledCommits gives the business logic access to information about the newly pulled commits.
	//
	// The list is available until the next call to Pull, which clears it and adds any new commits.
	NewPulledCommits() []CommitInfo

	// IsCommitKnown is true if the given commit has been cloned, pulled or locally committed, meaning,
	// a Pull would not generate new information if this commit hash is in the pull.
	IsCommitKnown(hash string) bool

	// standard git-aware file operations on the current worktree

	Stat(filename string) (os.FileInfo, error)
	ReadDir(path string) ([]os.FileInfo, error)

	// ReadFile returns the contents of a file, the commit hash, timestamp and message for the last change to the file
	ReadFile(filename string) ([]byte, CommitInfo, error)

	// WriteFile creates or overwrites a file in the local copy
	WriteFile(filename string, contents []byte) error

	// DeleteFile deletes a file in the local copy
	DeleteFile(filename string) error

	// Mkdir creates a new directory (and potentially all directories leading up to it). Does nothing if already exists.
	MkdirAll(path string) error
}
