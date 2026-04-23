package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/YourSurpriseCom/oapi-codegen-client/_test/cataas"
	"github.com/YourSurpriseCom/oapi-codegen-client/oapiclient"
)

func ExampleNew() {
	baseUrl := "https://cataas.com"
	upstreamTimeout := 5 * time.Second
	client := oapiclient.New[cataas.Client, cataas.ClientWithResponses](baseUrl, upstreamTimeout)

	response, err := client.CatRandomTextWithResponse(context.Background(), "test", &cataas.CatRandomTextParams{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", response.JSON200)
}
