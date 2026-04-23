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

	var doer HttpRequestDoer = httpClient
	if cfg.httpDoer != nil {
		doer = cfg.httpDoer
	}

	var client Client

	v := reflect.ValueOf(&client).Elem()

	if f := v.FieldByName("Server"); f.IsValid() && f.CanSet() {
		f.SetString(baseURL)
	}

	if f := v.FieldByName("Client"); f.IsValid() && f.CanSet() {
		f.Set(reflect.ValueOf(doer))
	}

	if cfg.oauthMiddleware != nil {
		if f := v.FieldByName("RequestEditors"); f.IsValid() && f.CanSet() {
			slice := reflect.MakeSlice(f.Type(), 1, 1)
			elemType := f.Type().Elem()
			slice.Index(0).Set(reflect.ValueOf(cfg.oauthMiddleware).Convert(elemType))
			f.Set(slice)
		}
	}

	var client2 ClientWithResponses
	v2 := reflect.ValueOf(&client2).Elem()
	if f2 := v2.FieldByName("ClientInterface"); f2.IsValid() && f2.CanSet() {
		f2.Set(reflect.ValueOf(&client))
	}

	return client2
}
