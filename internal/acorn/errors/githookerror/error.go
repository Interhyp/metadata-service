package githookerror

import "strings"

const hookError = "pre-receive hook declined"

// Is checks that an error is a git hook error.
//
// These errors occur during push operations.
//
// Unfortunately, we have to decide this by the error message, as go-git just uses fmt.Errorf().
func Is(err error) bool {
	return err != nil && strings.Contains(err.Error(), hookError)
}
