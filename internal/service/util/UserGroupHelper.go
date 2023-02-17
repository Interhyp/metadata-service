package util

import (
	"strings"
)

func ParseGroupOwnerAndGroupName(mayBeGroupReference string) (bool, string, string) {
	hasGroupPrefix := strings.HasPrefix(mayBeGroupReference, "@")
	indexOfDot := strings.Index(mayBeGroupReference, ".")
	if hasGroupPrefix && indexOfDot > 0 {
		return true, mayBeGroupReference[1:indexOfDot], mayBeGroupReference[indexOfDot+1:]
	}
	return false, "", ""
}

func SplitUsersAndGroups(userAndGroups []string) ([]string, []string) {
	users := make([]string, 0)
	groups := make([]string, 0)
	for _, userOrGroup := range userAndGroups {
		isGroup, _, _ := ParseGroupOwnerAndGroupName(userOrGroup)
		if isGroup {
			groups = append(groups, userOrGroup)
		} else {
			users = append(users, userOrGroup)
		}
	}
	return users, groups
}

func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func Difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
