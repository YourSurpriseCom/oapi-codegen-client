package gcp

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/idtoken"
)

func OauthMiddleware(audience string) (func(ctx context.Context, req *http.Request) error, error) {
	if audience == "" {
		return nil, fmt.Errorf("audience must not be empty")
	}

	tokenSource, err := idtoken.NewTokenSource(context.Background(), audience)
	if err != nil {
		return nil, fmt.Errorf("unexpected error creating token source: %w", err)
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
