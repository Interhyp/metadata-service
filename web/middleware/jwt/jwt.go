package jwt

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"github.com/Interhyp/metadata-service/acorns/config"
	"github.com/Interhyp/metadata-service/web/util"
	"github.com/go-http-utils/headers"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strings"
	"time"
)

type ctxJwtKeyType int

const (
	RawTokenKey ctxJwtKeyType = 0
	ClaimsKey   ctxJwtKeyType = 1
)

type CustomClaims struct {
	Name   string   `json:"name"`
	Email  string   `json:"email"`
	Groups []string `json:"groups"`
}

// end example

type AllClaims struct {
	// maybe * ?
	jwt.RegisteredClaims
	CustomClaims
}

var RsaPublicKeys = make([]*rsa.PublicKey, 0)
var customConfig config.CustomConfiguration

// Now exported for testing
var Now = time.Now

func Setup(publicKeyPEMs []string, config config.CustomConfiguration) error {
	for _, publicKeyPEM := range publicKeyPEMs {
		publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
		if err != nil {
			return err
		}

		RsaPublicKeys = append(RsaPublicKeys, publicKey)
	}

	customConfig = config
	return nil
}

func keyFuncForKey(rsaPublicKey *rsa.PublicKey) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		return rsaPublicKey, nil
	}
}

func JwtValidator(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		authHeaderValue := r.Header.Get(headers.Authorization)
		if authHeaderValue == "" {
			// valid case, no authorization provided
			next.ServeHTTP(w, r)
		} else {
			ctx := r.Context()
			username, password, basicAuthOk := r.BasicAuth()
			if basicAuthOk {
				if checkBasicAuthValue(username, password) {
					adminClaims := AllClaims{
						RegisteredClaims: jwt.RegisteredClaims{},
						CustomClaims: CustomClaims{
							Name:   customConfig.GitCommitterName(),
							Email:  customConfig.GitCommitterEmail(),
							Groups: strings.Fields(customConfig.AuthGroupWrite()),
						},
					}
					ctx = PutClaims(ctx, &adminClaims)
					next.ServeHTTP(w, r.WithContext(ctx))
				} else {
					util.UnauthorizedErrorHandler(ctx, w, r, "value of Authorization Basic header contains invalid values", Now())
					return
				}
			} else {
				const BearerPrefix = "Bearer "
				if !strings.HasPrefix(authHeaderValue, BearerPrefix) {
					util.UnauthorizedErrorHandler(ctx, w, r, "value of Authorization header did not start with 'Bearer '", Now())
					return
				}
				tokenString := strings.TrimSpace(strings.TrimPrefix(authHeaderValue, BearerPrefix))

				errorMessage := ""
				for _, key := range RsaPublicKeys {
					claims := AllClaims{}
					token, err := jwt.ParseWithClaims(tokenString, &claims, keyFuncForKey(key), jwt.WithValidMethods([]string{"RS256"}))
					if err == nil && token.Valid {
						parsedClaims := token.Claims.(*AllClaims)

						ctx = PutRawToken(ctx, token.Raw)
						ctx = PutClaims(ctx, parsedClaims)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
					if err != nil {
						errorMessage = err.Error()
					}
					if !token.Valid {
						errorMessage = "token parsed but invalid"
					}
				}
				util.UnauthorizedErrorHandler(ctx, w, r, errorMessage, Now())
			}
		}
	}
	return http.HandlerFunc(fn)
}

func checkBasicAuthValue(username string, password string) bool {
	expectedUsernameHash := sha256.Sum256([]byte(customConfig.BasicAuthUsername()))
	expectedPasswordHash := sha256.Sum256([]byte(customConfig.BasicAuthPassword()))

	usernameHash := sha256.Sum256([]byte(username))
	passwordHash := sha256.Sum256([]byte(password))

	usernameMatch := subtle.ConstantTimeCompare(expectedUsernameHash[:], usernameHash[:]) == 1
	passwordMatch := subtle.ConstantTimeCompare(expectedPasswordHash[:], passwordHash[:]) == 1

	return usernameMatch && passwordMatch
}

// GetRawToken returns the raw token from the given context if one is present.
//
// Returns the empty string if the context contains no valid token.
func GetRawToken(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if token, ok := ctx.Value(RawTokenKey).(string); ok {
		return token
	}
	return ""
}

// PutRawToken places a raw token in the context under the correct key.
//
// Returns a child context with the token set.
//
// Exposed for testing only.
func PutRawToken(ctx context.Context, rawToken string) context.Context {
	return context.WithValue(ctx, RawTokenKey, rawToken)
}

// GetClaims returns the raw token from the given context if one is present.
//
// Returns the empty string if the context contains no valid token.
func GetClaims(ctx context.Context) *AllClaims {
	if ctx == nil {
		return nil
	}
	if claimsPtr, ok := ctx.Value(ClaimsKey).(*AllClaims); ok {
		return claimsPtr
	}
	return nil
}

// PutClaims places a raw token in the context under the correct key.
//
// Returns a child context with the token set.
//
// Exposed for testing only.
func PutClaims(ctx context.Context, claimsPtr *AllClaims) context.Context {
	return context.WithValue(ctx, ClaimsKey, claimsPtr)
}

func IsAuthenticated(ctx context.Context) bool {
	claimsPtr := GetClaims(ctx)
	return claimsPtr != nil
}

func HasGroup(ctx context.Context, group string) bool {
	if group == "" {
		return true
	}
	claimsPtr := GetClaims(ctx)
	if claimsPtr == nil {
		return false
	}
	return contains(claimsPtr.Groups, group)
}

func Name(ctx context.Context) string {
	claimsPtr := GetClaims(ctx)
	if claimsPtr == nil {
		return ""
	}
	return claimsPtr.Name
}

func Email(ctx context.Context) string {
	claimsPtr := GetClaims(ctx)
	if claimsPtr == nil {
		return ""
	}
	return claimsPtr.Email
}

func Subject(ctx context.Context) string {
	claimsPtr := GetClaims(ctx)
	if claimsPtr == nil {
		return ""
	}
	return claimsPtr.RegisteredClaims.Subject
}

func contains(haystack []string, needle string) bool {
	if len(haystack) == 0 {
		return false
	}
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}
