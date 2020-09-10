package main

import (
	"context"
	"fmt"

	"github.com/listspa/go-artifactory/v2/artifactory"
	"github.com/listspa/go-artifactory/v2/artifactory/transport"
)

func main() {
	searchTemplate := `items.find({"repo": "example-repo-local","path": {"$ne": "."},"$or": [{"$and":[{"path": {"$match": "*"},"name": {"$match": "TEST"}}]}]}).include("name","repo","path","actual_md5","actual_sha1","size","type","property")`

	tp := transport.BasicAuth{
		Username: "admin",
		Password: "password",
	}

	rt, err := artifactory.NewClient("http://localhost:8091/artifactory", tp.Client(), "trace")
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	_, _, err = rt.V1.System.Ping(context.Background())
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	} else {
		fmt.Println("OK")
	}
	_, _, err = rt.V1.Artifacts.SearchByAQL(context.Background(), searchTemplate)
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	} else {
		fmt.Println("OK")
	}
}
