package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/YourSurpriseCom/oapi-codegen-client/_test/cataas"
	"github.com/YourSurpriseCom/oapi-codegen-client/oapiclient"
)

func ExampleNew() {
	baseURL := "https://cataas.com"
	upstreamTimeout := 5 * time.Second
	client := oapiclient.New[cataas.Client, cataas.ClientWithResponses](baseURL, upstreamTimeout)

	response, err := client.CatRandomTextWithResponse(context.Background(), "test", &cataas.CatRandomTextParams{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", response.StatusCode())
	// Output: 200
}
