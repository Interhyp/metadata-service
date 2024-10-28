package recorder

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
)

func ConstructFilenameV4(method string, requestUrl string, _ interface{}) (string, error) {
	parsedUrl, err := url.Parse(requestUrl)
	if err != nil {
		return "", err
	}

	m := strings.ToLower(method)
	md5sumOverPath := md5.Sum([]byte(parsedUrl.EscapedPath()))
	p := hex.EncodeToString(md5sumOverPath[:])
	p = p[:8]
	// we have to ensure the filenames don't get too long. git for windows only supports 260 character paths
	md5sumOverQuery := md5.Sum([]byte(parsedUrl.Query().Encode()))
	q := hex.EncodeToString(md5sumOverQuery[:])
	q = q[:8]

	filename := fmt.Sprintf("request_%s_%s_%s.json", m, p, q)
	return filename, nil
}
