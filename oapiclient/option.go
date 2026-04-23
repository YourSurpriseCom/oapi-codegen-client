package oapiclient

import (
	"context"
	"net/http"

	"github.com/YourSurpriseCom/go-datadog-apm/v2/apm"
	"github.com/YourSurpriseCom/oapi-codegen-client/internal/gcp"
)

// HttpRequestDoer is the interface for the HTTP client used internally by generated clients.
type HttpRequestDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type clientConfig struct {
	apm             *apm.Apm
	oauthMiddleware func(context.Context, *http.Request) error
	httpDoer        HttpRequestDoer
}

// WithHttpDoer overrides the HTTP transport, useful for injecting mocks in tests.
func WithHttpDoer(doer HttpRequestDoer) ClientOption {
	return func(config *clientConfig) {
		config.httpDoer = doer
	}
}

type ClientOption func(*clientConfig)

// WithDatadogApm enables DataDog http trace client inside the http client
func WithDatadogApm(apm *apm.Apm) ClientOption {
	return func(clientConfig *clientConfig) {
		clientConfig.apm = apm
	}
}

// WithGcpOAuth enables Google Cloud Platform authentication
func WithGcpOAuth(audience string) ClientOption {
	middleware, err := gcp.OauthMiddleware(audience)
	if err != nil {
		panic(err)
	}

	return func(config *clientConfig) { config.oauthMiddleware = middleware }
}
