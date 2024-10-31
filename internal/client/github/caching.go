package githubclient

import (
	"github.com/gregjones/httpcache"
	"net/http"
)

type CacheConditionCallback func(method string, url string, requestBody interface{}) bool

type Caching struct {
	Wrapped           *http.Client
	Cache             *httpcache.Transport
	UseCacheCondition CacheConditionCallback
}

func NewCaching(
	wrapped *http.Client,
	useCacheCondition CacheConditionCallback,
) *Caching {
	if wrapped == nil {
		wrapped = http.DefaultClient
		wrapped.Transport = http.DefaultTransport
	}
	return &Caching{
		Wrapped:           wrapped,
		Cache:             httpcache.NewMemoryCacheTransport(),
		UseCacheCondition: useCacheCondition,
	}
}

func (c *Caching) Client() *http.Client {
	return &http.Client{Transport: c}
}

func (c *Caching) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	canCache := c.UseCacheCondition(req.Method, req.URL.EscapedPath(), req.Body)
	if canCache {
		return c.Cache.RoundTrip(req)
	}
	return c.Wrapped.Transport.RoundTrip(req)
}
