package oapiclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/YourSurpriseCom/go-datadog-apm/v2/apm"
	"github.com/YourSurpriseCom/oapi-codegen-client/_test/cataas"
	"github.com/stretchr/testify/require"
)

func noopOauthMiddleware(_ string) ClientOption {
	return func(config *clientConfig) {
		config.oauthMiddleware = func(ctx context.Context, req *http.Request) error { return nil }
	}
}

func failingOauthMiddleware(_ string) ClientOption {
	return func(config *clientConfig) {
		config.oauthMiddleware = func(ctx context.Context, req *http.Request) error { return errors.New("token error") }
	}
}

func TestNew(t *testing.T) {
	baseUrl := "https://cataas.com"
	upstreamTimeout := 5 * time.Second

	tests := []struct {
		name          string
		options       []ClientOption
		expectedError string
		expectedCode  int
	}{
		{
			name:         "no middleware",
			options:      nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "noop oauth middleware",
			options:      []ClientOption{noopOauthMiddleware("test-audience")},
			expectedCode: http.StatusOK,
		},
		{
			name:          "failing oauth middleware",
			options:       []ClientOption{failingOauthMiddleware("test-audience")},
			expectedError: "token error",
		},
		{
			name: "with datadog apm",
			options: []ClientOption{func() ClientOption {
				a := apm.NewApm()
				return WithDatadogApm(&a)
			}()},
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New[cataas.Client, cataas.ClientWithResponses](baseUrl, upstreamTimeout, tt.options...)

			// apply the interface to validate the client
			var _ cataas.ClientWithResponsesInterface = &client

			response, err := client.CatRandomTextWithResponse(context.Background(), "test", &cataas.CatRandomTextParams{})

			if tt.expectedError != "" {
				require.ErrorContains(t, err, tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expectedCode, response.StatusCode())

			fmt.Printf("%+v\n", response.JSON200)
		})
	}
}

func TestWithGcpOAuth_PanicsOnEmptyAudience(t *testing.T) {
	require.Panics(t, func() {
		WithGcpOAuth("")
	})
}
