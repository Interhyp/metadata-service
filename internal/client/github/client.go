package githubclient

import (
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/google/go-github/v66/github"
	"net/http"
)

func NewClient(client *http.Client, accessToken string, customConfig config.CustomConfiguration) (*github.Client, error) {
	return github.NewClient(client).WithAuthToken(accessToken), nil
}
