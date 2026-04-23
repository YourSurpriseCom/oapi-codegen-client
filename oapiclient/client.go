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

	if serverField := clientValue.FieldByName("Server"); serverField.IsValid() && serverField.CanSet() {
		serverField.SetString(baseURL)
	}

	if clientField := clientValue.FieldByName("Client"); clientField.IsValid() && clientField.CanSet() {
		clientField.Set(reflect.ValueOf(doer))
	}

	if cfg.oauthMiddleware != nil {
		if editorsField := clientValue.FieldByName("RequestEditors"); editorsField.IsValid() && editorsField.CanSet() {
			slice := reflect.MakeSlice(editorsField.Type(), 1, 1)
			elemType := editorsField.Type().Elem()
			slice.Index(0).Set(reflect.ValueOf(cfg.oauthMiddleware).Convert(elemType))
			editorsField.Set(slice)
		}
	}

	var clientWithResponses ClientWithResponses
	clientWithResponsesValue := reflect.ValueOf(&clientWithResponses).Elem()
	if clientInterfaceField := clientWithResponsesValue.FieldByName("ClientInterface"); clientInterfaceField.IsValid() && clientInterfaceField.CanSet() {
		clientInterfaceField.Set(reflect.ValueOf(&client))
	}

	return clientWithResponses
}
