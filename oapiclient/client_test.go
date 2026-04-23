package oapiclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/YourSurpriseCom/go-datadog-apm/v2/apm"
	mockcataas "github.com/YourSurpriseCom/oapi-codegen-client/_mocks/cataas"
	"github.com/YourSurpriseCom/oapi-codegen-client/_test/cataas"
	"github.com/stretchr/testify/mock"
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

var defaultTestResponseSuccessURL = "https://example.com/cat"
var defaultTestResponseSuccess = cataas.CatRandomTextResponse{
	HTTPResponse: &http.Response{
		StatusCode: http.StatusOK,
	},
	JSON200: &cataas.Cat{
		Url: &defaultTestResponseSuccessURL,
	},
}

func successHTTPResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"url":"https://example.com/cat"}`)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func TestNew(t *testing.T) {
	baseURL := "https://cataas.com"
	upstreamTimeout := 5 * time.Second

	tests := []struct {
		name          string
		options       []ClientOption
		expectedError string
		expectedCode  int
		mockResponse  *http.Response
	}{
		{
			name:         "no middleware",
			options:      nil,
			expectedCode: http.StatusOK,
			mockResponse: successHTTPResponse(),
		},
		{
			name:         "noop oauth middleware",
			options:      []ClientOption{noopOauthMiddleware("test-audience")},
			expectedCode: http.StatusOK,
			mockResponse: successHTTPResponse(),
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
			mockResponse: successHTTPResponse(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDoer := mockcataas.NewMockHttpRequestDoer(t)
			if tt.mockResponse != nil {
				mockDoer.EXPECT().Do(mock.Anything).Return(tt.mockResponse, nil)
			}

			tt.options = append(tt.options, WithHTTPDoer(mockDoer))
			client := New[cataas.Client, cataas.ClientWithResponses](baseURL, upstreamTimeout, tt.options...)

			var _ cataas.ClientWithResponsesInterface = &client

			response, err := client.CatRandomTextWithResponse(context.Background(), "test", &cataas.CatRandomTextParams{})

			if tt.expectedError != "" {
				require.ErrorContains(t, err, tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expectedCode, response.StatusCode())
			require.Equal(t, defaultTestResponseSuccess.JSON200, response.JSON200)
		})
	}
}

func TestWithGcpOAuthPanicsOnEmptyAudience(t *testing.T) {
	require.Panics(t, func() {
		WithGcpOAuth("")
	})
}
