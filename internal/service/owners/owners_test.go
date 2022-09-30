package owners

import (
	openapi "github.com/Interhyp/metadata-service/api/v1"
	"github.com/Interhyp/metadata-service/docs"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPatchOwner(t *testing.T) {
	docs.Description("patching of owners works")

	productowner := "productowner"
	jiraproject := "jira"

	newProductowner := "productowner.new"
	newJiraproject := "jira.new"
	newContact := "contact.new"

	emptyJiraproject := ""

	current := openapi.OwnerDto{
		Contact:            "contact",
		ProductOwner:       &productowner,
		DefaultJiraProject: &jiraproject,
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
		TimeStamp:          "timestamp.new",
		CommitHash:         "commithash.new",
	}
	require.Equal(t, expected0, actual0)

	patch1 := openapi.OwnerPatchDto{
		Contact:    &newContact,
		TimeStamp:  "timestamp.new",
		CommitHash: "commithash.new",
	}
	actual1 := patchOwner(current, patch1)
	expected1 := openapi.OwnerDto{
		Contact:            "contact.new",
		ProductOwner:       &productowner,
		DefaultJiraProject: &jiraproject,
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
		TimeStamp:          "timestamp.new",
		CommitHash:         "commithash.new",
	}
	require.Equal(t, expected3, actual3)
}
