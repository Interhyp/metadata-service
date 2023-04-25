package acceptance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"os"
	"testing"
)

// placing these here because they are package global

func tstAssertNoBody(t *testing.T, response tstWebResponse, err error, expectedStatus int) {
	require.Nil(t, err)
	require.Equal(t, expectedStatus, response.status)
	require.Equal(t, "", response.body)
}

func tstAssert(t *testing.T, response tstWebResponse, err error, expectedStatus int, expectedBodyFile string) {
	require.Nil(t, err)
	require.Equal(t, expectedStatus, response.status)
	require.Equal(t, "application/json", response.contentType)

	if os.Getenv("REPLACE_TEST_RECORDINGS") != "" {
		// write recorded value if environment variable set
		bytes := tstPrettyprintJsonObject(response.body)
		_ = os.WriteFile(fmt.Sprintf("../resources/acceptance-expected/%s", expectedBodyFile), bytes, 0644)
	}

	expectedBody := tstReadExpected(expectedBodyFile)

	expectedJsonNormalized := tstUnprettyprintJsonObject(expectedBody)
	actualJsonNormalized := tstUnprettyprintJsonObject(response.body)
	require.Equal(t, expectedJsonNormalized, actualJsonNormalized)
	require.NotEqual(t, "unprettyprint error", actualJsonNormalized)
}

func tstReadExpected(filename string) string {
	content, err := os.ReadFile(fmt.Sprintf("../resources/acceptance-expected/%s", filename))
	if err != nil {
		return fmt.Sprintf("error reading file %s", filename)
	}
	return string(content)
}

func tstUnprettyprintJsonObject(prettyprinted string) string {
	tmp := make(map[string]interface{})
	err := json.Unmarshal([]byte(prettyprinted), &tmp)
	if err != nil {
		return "unprettyprint error"
	}
	result, err := json.Marshal(&tmp)
	if err != nil {
		return "unprettyprint error"
	}
	return string(result)
}

func tstPrettyprintJsonObject(prettyprinted string) []byte {
	tmp := make(map[string]interface{})
	err := json.Unmarshal([]byte(prettyprinted), &tmp)
	if err != nil {
		return []byte("prettyprint error")
	}
	result, err := json.MarshalIndent(&tmp, "", "  ")
	if err != nil {
		return []byte("prettyprint error")
	}
	return result
}

type tstWebResponse struct {
	status      int
	body        string
	contentType string
	location    string
}

func tstWebResponseFromResponse(response *http.Response) (tstWebResponse, error) {
	status := response.StatusCode
	ct := ""
	if val, ok := response.Header[headers.ContentType]; ok {
		ct = val[0]
	}
	loc := ""
	if val, ok := response.Header[headers.Location]; ok {
		loc = val[0]
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return tstWebResponse{}, err
	}
	err = response.Body.Close()
	if err != nil {
		return tstWebResponse{}, err
	}
	return tstWebResponse{
		status:      status,
		body:        string(body),
		contentType: ct,
		location:    loc,
	}, nil
}

func tstPerformGet(relativeUrlWithLeadingSlash string, bearerToken string) (tstWebResponse, error) {
	return tstPerformNoBody(http.MethodGet, relativeUrlWithLeadingSlash, bearerToken)
}

func tstPerformDeleteNoBody(relativeUrlWithLeadingSlash string, bearerToken string) (tstWebResponse, error) {
	return tstPerformNoBody(http.MethodDelete, relativeUrlWithLeadingSlash, bearerToken)
}

func tstPerformNoBody(method string, relativeUrlWithLeadingSlash string, bearerToken string) (tstWebResponse, error) {
	if ts == nil {
		return tstWebResponse{}, errors.New("test web server was not initialized")
	}
	request, err := http.NewRequest(method, ts.URL+relativeUrlWithLeadingSlash, nil)
	if err != nil {
		return tstWebResponse{}, err
	}
	if bearerToken != "" {
		request.Header.Set(headers.Authorization, "Bearer "+bearerToken)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return tstWebResponse{}, err
	}
	return tstWebResponseFromResponse(response)
}

func tstPerformPost(relativeUrlWithLeadingSlash string, bearerToken string, bodyPtr interface{}) (tstWebResponse, error) {
	return tstPerformWithBody(http.MethodPost, relativeUrlWithLeadingSlash, bearerToken, bodyPtr)
}

func tstPerformPut(relativeUrlWithLeadingSlash string, bearerToken string, bodyPtr interface{}) (tstWebResponse, error) {
	return tstPerformWithBody(http.MethodPut, relativeUrlWithLeadingSlash, bearerToken, bodyPtr)
}

func tstPerformPatch(relativeUrlWithLeadingSlash string, bearerToken string, bodyPtr interface{}) (tstWebResponse, error) {
	return tstPerformWithBody(http.MethodPatch, relativeUrlWithLeadingSlash, bearerToken, bodyPtr)
}

func tstPerformDelete(relativeUrlWithLeadingSlash string, bearerToken string, bodyPtr interface{}) (tstWebResponse, error) {
	return tstPerformWithBody(http.MethodDelete, relativeUrlWithLeadingSlash, bearerToken, bodyPtr)
}

func tstPerformWithBody(method string, relativeUrlWithLeadingSlash string, bearerToken string, bodyPtr interface{}) (tstWebResponse, error) {
	bodyBytes, err := json.Marshal(bodyPtr)
	if err != nil {
		return tstWebResponse{}, err
	}
	return tstPerformRawWithBody(method, relativeUrlWithLeadingSlash, bearerToken, bodyBytes)
}

func tstPerformRawWithBody(method string, relativeUrlWithLeadingSlash string, bearerToken string, bodyBytes []byte) (tstWebResponse, error) {
	if ts == nil {
		return tstWebResponse{}, errors.New("test web server was not initialized")
	}
	request, err := http.NewRequest(method, ts.URL+relativeUrlWithLeadingSlash, bytes.NewReader(bodyBytes))
	if err != nil {
		return tstWebResponse{}, err
	}
	if bearerToken != "" {
		request.Header.Set(headers.Authorization, "Bearer "+bearerToken)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return tstWebResponse{}, err
	}
	return tstWebResponseFromResponse(response)
}
