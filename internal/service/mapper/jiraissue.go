package mapper

import (
	"regexp"
	"strings"
)

const mergeMsgRegex = "Pull request #[0-9]+\\: "

func jiraIssue(commitMessage string) string {
	regex, _ := regexp.Compile(mergeMsgRegex)
	commitMessage = regex.ReplaceAllString(commitMessage, "")

	fields := strings.FieldsFunc(commitMessage, func(r rune) bool {
		// split at anything that is not A-Z 0-9 -
		if 'A' <= r && r <= 'Z' {
			return false
		} else if '0' <= r && r <= '9' {
			return false
		} else if '-' == r {
			return false
		} else {
			return true
		}
	})
	if len(fields) > 0 && len(fields[0]) > 1 {
		return fields[0]
	} else {
		return ""
	}
}
