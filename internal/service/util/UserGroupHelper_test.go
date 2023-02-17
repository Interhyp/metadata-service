package util

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseGroupOwnerAndGroupName(t *testing.T) {

	isGroup, ownerOfGroup, nameOfGroup := ParseGroupOwnerAndGroupName("@someOwner.someGroupName")
	require.True(t, isGroup)
	require.Equal(t, "someOwner", ownerOfGroup)
	require.Equal(t, "someGroupName", nameOfGroup)

	isGroup, ownerOfGroup, nameOfGroup = ParseGroupOwnerAndGroupName("someOwner.someGroupName")
	require.False(t, isGroup)
	require.Equal(t, "", ownerOfGroup)
	require.Equal(t, "", nameOfGroup)

	isGroup, ownerOfGroup, nameOfGroup = ParseGroupOwnerAndGroupName("@someGroupName")
	require.False(t, isGroup)
	require.Equal(t, "", ownerOfGroup)
	require.Equal(t, "", nameOfGroup)
}
