package owners

import (
	"context"
	openapi "github.com/Interhyp/metadata-service/api/v1"
	"github.com/Interhyp/metadata-service/docs"
	ownersmock "github.com/Interhyp/metadata-service/test/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetAllGroupMembers(t *testing.T) {
	instance := Impl{
		Configuration: nil,
		Logging:       nil,
		Cache:         &ownersmock.Mock{},
		Updater:       nil,
	}
	groupMembers := instance.GetAllGroupMembers(context.Background(), "ownerWithGroup", "someGroupName")
	require.Equal(t, 2, len(groupMembers))
	require.Contains(t, groupMembers, "username1")
	require.Contains(t, groupMembers, "username2")

	groupMembers = instance.GetAllGroupMembers(context.Background(), "someOwner", "someGroupName")
	require.Equal(t, 0, len(groupMembers))
}

func TestPatchOwner(t *testing.T) {
	docs.Description("patching of owners works")

	productowner := "productowner"
	jiraproject := "jira"
	groups := map[string][]string{
		"group0": {"user1", "user2"},
	}

	newProductowner := "productowner.new"
	newJiraproject := "jira.new"
	newContact := "contact.new"
	newGroups := map[string][]string{
		"group1": {"user1"},
	}

	emptyJiraproject := ""

	current := openapi.OwnerDto{
		Contact:            "contact",
		ProductOwner:       &productowner,
		DefaultJiraProject: &jiraproject,
		Groups:             &groups,
		TimeStamp:          "timestamp",
		CommitHash:         "commithash",
	}

	patch0 := openapi.OwnerPatchDto{
		TimeStamp:  "timestamp.new",
		CommitHash: "commithash.new",
	}
	actual0 := patchOwner(current, patch0)
	expected0 := openapi.OwnerDto{
		Contact:            "contact",
		ProductOwner:       &productowner,
		DefaultJiraProject: &jiraproject,
		Groups:             &groups,
		TimeStamp:          "timestamp.new",
		CommitHash:         "commithash.new",
	}
	require.Equal(t, expected0, actual0)

	patch1 := openapi.OwnerPatchDto{
		Contact:    &newContact,
		TimeStamp:  "timestamp.new",
		CommitHash: "commithash.new",
		Groups:     &newGroups,
	}
	actual1 := patchOwner(current, patch1)
	expected1 := openapi.OwnerDto{
		Contact:            "contact.new",
		ProductOwner:       &productowner,
		DefaultJiraProject: &jiraproject,
		Groups:             &newGroups,
		TimeStamp:          "timestamp.new",
		CommitHash:         "commithash.new",
	}
	require.Equal(t, expected1, actual1)

	patch2 := openapi.OwnerPatchDto{
		ProductOwner:       &newProductowner,
		DefaultJiraProject: &emptyJiraproject,
		TimeStamp:          "timestamp.new",
		CommitHash:         "commithash.new",
	}
	actual2 := patchOwner(current, patch2)
	expected2 := openapi.OwnerDto{
		Contact:            "contact",
		ProductOwner:       &newProductowner,
		DefaultJiraProject: nil,
		Groups:             &groups,
		TimeStamp:          "timestamp.new",
		CommitHash:         "commithash.new",
	}
	require.Equal(t, expected2, actual2)

	patch3 := openapi.OwnerPatchDto{
		DefaultJiraProject: &newJiraproject,
		TimeStamp:          "timestamp.new",
		CommitHash:         "commithash.new",
	}
	actual3 := patchOwner(current, patch3)
	expected3 := openapi.OwnerDto{
		Contact:            "contact",
		ProductOwner:       &productowner,
		DefaultJiraProject: &newJiraproject,
		Groups:             &groups,
		TimeStamp:          "timestamp.new",
		CommitHash:         "commithash.new",
	}
	require.Equal(t, expected3, actual3)
}
