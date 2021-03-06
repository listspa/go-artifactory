package main

import (
	"context"
	"fmt"
	"os"

	"github.com/listspa/go-artifactory/v2/artifactory"
	"github.com/listspa/go-artifactory/v2/artifactory/transport"
)

func main() {
	tp := transport.ApiKeyAuth{
		ApiKey: os.Getenv("ARTIFACTORY_API_KEY"),
	}

	client, err := artifactory.NewClient(os.Getenv("ARTIFACTORY_URL"), tp.Client(), "debug")
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	_, _, err = client.V1.System.Ping(context.Background())
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
	} else {
		fmt.Println("OK")
	}
}
