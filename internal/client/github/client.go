package githubclient

import (
	"github.com/google/go-github/v66/github"
	"net/http"
)

func NewClient(client *http.Client, accessToken string) (*github.Client, error) {
	return github.NewClient(client).WithAuthToken(accessToken), nil
}
