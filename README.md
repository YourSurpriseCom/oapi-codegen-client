# Oapi-codegen client
A Golang client wrapper for clients generated with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen/).

# Description
This package wraps around the generated clients from [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen/). By using this package it lowers the boilerplate code for adding Datadog APM or Google Cloud Platform OAuth support.

## Usage
`go get github.com/YourSurpriseCom/oapi-codegen-client`

See [examples/example.go](examples/example.go) for a full usage example.

## Configuration
This client support multiple configurations which can be used on top of the genereted client.

### Datadog Support
To enable the Datadog http tracer, use the option `WithDatadogApm()` which expects an instance of [go-datadog-apm](https://github.com/YourSurpriseCom/go-datadog-apm).

### Google Cloud Platform OAuth support
To enable the Google Cloud Platform Oauth, use the option `WithGcpOAuth()` which expects the audience as a param.