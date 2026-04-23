# Oapi-codegen client
A Golang client wrapper for clients generated with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen/).

# Description
This package wraps around the generated clients from [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen/). By using this package it lowers the boilerplate code for adding Datadog APM or Google Cloud Platform OAuth support.

## Usage
`go get github.com/YourSurpriseCom/oapi-codegen-client`

```go 
baseURL := "https://cataas.com"
upstreamTimeout := 5 * time.Second
apmInstance := apm.NewApm()
audience := "https://audience"

// default client without extras
client := oapiclient.New[cataas.Client, cataas.ClientWithResponses](baseURL, upstreamTimeout)

// Client with Datadog instrumentation
client := oapiclient.New[cataas.Client, cataas.ClientWithResponses](baseURL, upstreamTimeout, oapiclient.WithDatadogApm(&apmInstance))

// Client with Google Cloud Platform authentication
client := oapiclient.New[cataas.Client, cataas.ClientWithResponses](baseURL, upstreamTimeout, oapiclient.WithGcpOAuth(audience))
```

See [examples/example_test.go](examples/example_test.go) for a full usage example.

## Configuration
This client support multiple configurations which can be used on top of the generated client.

### Datadog Support
To enable the Datadog http tracer, use the option `WithDatadogApm()` which expects an instance of [go-datadog-apm](https://github.com/YourSurpriseCom/go-datadog-apm).

### Google Cloud Platform OAuth support
To enable the Google Cloud Platform Oauth, use the option `WithGcpOAuth()` which expects the audience as a param.