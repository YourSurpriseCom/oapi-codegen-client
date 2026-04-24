package oapiclient

import (
	"net/http"
	"reflect"
	"time"
)

func New[Client any, ClientWithResponses any](baseURL string, upstreamTimeout time.Duration, clientOptions ...ClientOption) ClientWithResponses {
	cfg := &clientConfig{}
	for _, option := range clientOptions {
		option(cfg)
	}

	httpClient := &http.Client{
		Timeout: upstreamTimeout,
	}
	if cfg.apm != nil {
		httpClient = cfg.apm.ConfigureOnHttpClient(httpClient)
	}

	var doer HTTPRequestDoer = httpClient
	if cfg.httpDoer != nil {
		doer = cfg.httpDoer
	}

	var client Client

	clientValue := reflect.ValueOf(&client).Elem()

	serverField := clientValue.FieldByName("Server")
	if !serverField.IsValid() || !serverField.CanSet() {
		panic("Client must have a Server field containing the baseURL from the external service")
	}
	serverField.SetString(baseURL)

	clientField := clientValue.FieldByName("Client")
	if !clientField.IsValid() || !clientField.CanSet() {
		panic("Client must have a Client field supporting a httpClient")
	}
	clientField.Set(reflect.ValueOf(doer))

	if cfg.oauthMiddleware != nil {
		editorsField := clientValue.FieldByName("RequestEditors")
		if !editorsField.IsValid() || !editorsField.CanSet() {
			panic("Client must have a RequestEditors field supporting a RequestEditor callback function")
		}

		slice := reflect.MakeSlice(editorsField.Type(), 1, 1)
		elemType := editorsField.Type().Elem()
		slice.Index(0).Set(reflect.ValueOf(cfg.oauthMiddleware).Convert(elemType))
		editorsField.Set(slice)
	}

	var clientWithResponses ClientWithResponses
	clientWithResponsesValue := reflect.ValueOf(&clientWithResponses).Elem()
	clientInterfaceField := clientWithResponsesValue.FieldByName("ClientInterface")
	if !clientInterfaceField.IsValid() || !clientInterfaceField.CanSet() {
		panic("ClientWithResponses must have a ClientInterface field supporting a Client interface")
	}
	clientInterfaceField.Set(reflect.ValueOf(&client))

	return clientWithResponses
}
