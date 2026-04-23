package gcp

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func fakeServiceAccountJSON(t *testing.T, privateKey *rsa.PrivateKey, tokenURI string) string {
	t.Helper()
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	payload, err := json.Marshal(map[string]string{
		"type":           "service_account",
		"project_id":     "fake-project",
		"private_key_id": "fake-key-id",
		"private_key":    string(privateKeyPEM),
		"client_email":   "fake@fake-project.iam.gserviceaccount.com",
		"client_id":      "123456789",
		"auth_uri":       "https://accounts.google.com/o/oauth2/auth",
		"token_uri":      tokenURI,
	})
	require.NoError(t, err)
	return string(payload)
}

func fakeIDToken(t *testing.T, privateKey *rsa.PrivateKey) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT","kid":"fake-key-id"}`))
	payloadBytes, err := json.Marshal(map[string]any{
		"iss": "fake@fake-project.iam.gserviceaccount.com",
		"aud": "https://example.com",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"sub": "fake@fake-project.iam.gserviceaccount.com",
	})
	require.NoError(t, err)
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	signingInput := header + "." + payload
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	require.NoError(t, err)

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
}

type errorTokenSource struct{ err error }

func (source errorTokenSource) Token() (*oauth2.Token, error) { return nil, source.err }

func TestOauthMiddlewareWithFakeCredentials(t *testing.T) {
	t.Run("idtoken-source-success", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		tokenResponseJSON, err := json.Marshal(map[string]any{
			"access_token": "fake-access-token",
			"id_token":     fakeIDToken(t, privateKey),
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
		require.NoError(t, err)

		tokenServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
			responseWriter.Header().Set("Content-Type", "application/json")
			responseWriter.Write(tokenResponseJSON)
		}))
		defer tokenServer.Close()

		credFile := filepath.Join(t.TempDir(), "sa.json")
		require.NoError(t, os.WriteFile(credFile, []byte(fakeServiceAccountJSON(t, privateKey, tokenServer.URL)), 0600))
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credFile)

		middleware, err := OauthMiddleware("https://example.com")
		require.NoError(t, err)
		require.NotNil(t, middleware)
	})

}

func TestOauthMiddleware(t *testing.T) {
	tests := []struct {
		name             string
		audience         string
		opts             []Option
		credentials      string // GOOGLE_APPLICATION_CREDENTIALS override; empty means don't set
		expectedError    string
		invokeMiddleware bool
		middlewareError  string
		expectedHeader   string
	}{
		{
			name:          "empty audience",
			audience:      "",
			expectedError: "audience must not be empty",
		},
		{
			name:          "idtoken creation failure",
			audience:      "https://example.com",
			credentials:   "/dev/null",
			expectedError: "unexpected error creating token source",
		},
		{
			name:             "middleware: token fetch error",
			audience:         "https://example.com",
			opts:             []Option{WithTokenSource(errorTokenSource{err: errors.New("token fetch failed")})},
			invokeMiddleware: true,
			middlewareError:  "failed to get id token",
		},
		{
			name:     "middleware: success",
			audience: "https://example.com",
			opts: []Option{
				WithTokenSource(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"})),
			},
			invokeMiddleware: true,
			expectedHeader:   "Bearer test-token",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// always reset the Google Application Credentials before testing, with nothing or the to be tested value
			t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", testCase.credentials)

			middleware, err := OauthMiddleware(testCase.audience, testCase.opts...)

			if testCase.expectedError != "" {
				require.ErrorContains(t, err, testCase.expectedError)
				require.Nil(t, middleware)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, middleware)

			if testCase.invokeMiddleware {
				req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
				middlewareErr := middleware(context.Background(), req)

				if testCase.middlewareError != "" {
					require.ErrorContains(t, middlewareErr, testCase.middlewareError)
					return
				}

				require.NoError(t, middlewareErr)
				require.Equal(t, testCase.expectedHeader, req.Header.Get("Authorization"))
			}
		})
	}
}
