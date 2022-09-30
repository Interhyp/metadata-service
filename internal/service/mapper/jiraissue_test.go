package mapper

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestJiraIssue_EmptyCommitMessage(t *testing.T) {
	result := jiraIssue("")
	require.Equal(t, "", result)
}

func TestJiraIssue_DefaultCommitMessageStyle(t *testing.T) {
	result := jiraIssue("ISSUE-000: some commit message text")
	require.Equal(t, "ISSUE-000", result)
}

func TestJiraIssue_MergeCommitMessage(t *testing.T) {
	result := jiraIssue("Pull request #23: ISSUE-000: some commit message text")
	require.Equal(t, "ISSUE-000", result)
}
