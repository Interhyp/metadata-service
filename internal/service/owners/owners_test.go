package owners

import (
	"context"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/test/mock/cachemock"
	"github.com/StephanHCB/go-backend-service-common/docs"
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
func ptr(in string) *string {
	return &in
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
	newLinks := []openapi.Link{
		{
			Url:   ptr("www.heute.de"),
			Title: ptr("ZDF Heute Nachrichten"),
		},
	}
	currentLinks := []openapi.Link{
		{
			Url:   ptr("www.interhyp.de"),
			Title: ptr("Interhyp"),
		},
	}

	emptyJiraproject := ""

	current := openapi.OwnerDto{
		Contact:            "contact",
		ProductOwner:       &productowner,
		DefaultJiraProject: &jiraproject,
		Groups:             &groups,
		TimeStamp:          "timestamp",
		CommitHash:         "commithash",
		Links:              currentLinks,
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
		Links:              currentLinks,
	}
	require.Equal(t, expected0, actual0)

	patch1 := openapi.OwnerPatchDto{
		Contact:    &newContact,
		TimeStamp:  "timestamp.new",
		CommitHash: "commithash.new",
		Groups:     &newGroups,
		Links:      newLinks,
	}
	actual1 := patchOwner(current, patch1)
	expected1 := openapi.OwnerDto{
		Contact:            "contact.new",
		ProductOwner:       &productowner,
		DefaultJiraProject: &jiraproject,
		Groups:             &newGroups,
		TimeStamp:          "timestamp.new",
		CommitHash:         "commithash.new",
		Links:              newLinks,
	}
	require.Equal(t, expected1, actual1)

	patch2 := openapi.OwnerPatchDto{
		ProductOwner:       &newProductowner,
		DefaultJiraProject: &emptyJiraproject,
		TimeStamp:          "timestamp.new",
		CommitHash:         "commithash.new",
		Links:              make([]openapi.Link, 0),
	}
	actual2 := patchOwner(current, patch2)
	expected2 := openapi.OwnerDto{
		Contact:            "contact",
		ProductOwner:       &newProductowner,
		DefaultJiraProject: nil,
		Groups:             &groups,
		TimeStamp:          "timestamp.new",
		CommitHash:         "commithash.new",
		Links:              nil,
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
		Links:              currentLinks,
	}
	require.Equal(t, expected3, actual3)
}
