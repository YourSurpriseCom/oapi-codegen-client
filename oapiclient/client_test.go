package oapiclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/YourSurpriseCom/oapi-codegen-client/_test/cataas"
	"github.com/stretchr/testify/require"
)

func noopOauthMiddleware(audience string) ClientOption {
	return func(config *clientConfig) {
		config.oauthMiddleware = func(ctx context.Context, req *http.Request) error { return nil }
	}
}

func failingOauthMiddleware(audience string) ClientOption {
	return func(config *clientConfig) {
		config.oauthMiddleware = func(ctx context.Context, req *http.Request) error { return errors.New("token error") }
	}
}

func TestNew(t *testing.T) {
	baseUrl := "https://cataas.com"
	upstreamTimeout := 5 * time.Second
	client := New[cataas.Client, cataas.ClientWithResponses](baseUrl, upstreamTimeout)

	// apply the interface to validate the client
	var _ cataas.ClientWithResponsesInterface = &client

	response, err := client.CatRandomTextWithResponse(context.Background(), "test", &cataas.CatRandomTextParams{})

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode())

	fmt.Printf("%+v\n", response.JSON200)
}
