package gcp

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

type oauthMiddlewareConfig struct {
	tokenSource oauth2.TokenSource
}

type Option func(*oauthMiddlewareConfig)

// WithTokenSource injects a custom token source, bypassing GCP credential discovery.
// Primarily useful for testing.
func WithTokenSource(ts oauth2.TokenSource) Option {
	return func(cfg *oauthMiddlewareConfig) {
		cfg.tokenSource = ts
	}
}

func OauthMiddleware(audience string, options ...Option) (func(ctx context.Context, req *http.Request) error, error) {
	if audience == "" {
		return nil, fmt.Errorf("audience must not be empty")
	}

	cfg := &oauthMiddlewareConfig{}
	for _, option := range options {
		option(cfg)
	}

	tokenSource := cfg.tokenSource
	if tokenSource == nil {
		idTokenSource, err := idtoken.NewTokenSource(context.Background(), audience)
		if err != nil {
			return nil, fmt.Errorf("unexpected error creating token source: %w", err)
		}
		tokenSource = idTokenSource
	}

	return func(ctx context.Context, req *http.Request) error {
		token, tokenErr := tokenSource.Token()
		if tokenErr != nil {
			return fmt.Errorf("failed to get id token: %w", tokenErr)
		}
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
		return nil
	}, nil
}
