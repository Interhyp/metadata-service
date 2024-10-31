package githubclient

import (
	"fmt"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/google/go-github/v66/github"
	"net/http"
	"regexp"
)

func NewClient(client *http.Client, accessToken string, customConfig config.CustomConfiguration) (*github.Client, error) {
	matcher := regexp.MustCompile(fmt.Sprintf("/users/[a-z_-]*%s", customConfig.UserPrefix()))
	cacheClient := NewCaching(
		client,
		func(method string, url string, requestBody interface{}) bool {
			return method == http.MethodGet && matcher.MatchString(url)
		},
	).Client()
	return github.NewClient(cacheClient).WithAuthToken(accessToken), nil
}
