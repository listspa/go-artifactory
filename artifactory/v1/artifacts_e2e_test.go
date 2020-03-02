package v1

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/listspa/go-artifactory/v2/artifactory/client"
	"github.com/listspa/go-artifactory/v2/artifactory/transport"
	"github.com/stretchr/testify/assert"
)

func TestE2E(t *testing.T) {
	tp := transport.BasicAuth{
		Username: "admin",
		Password: "password",
	}
	c, _ := client.NewClient("http://localhost:8081/artifactory", tp.Client())
	v := NewV1(c)
	props := []ArtifactoryProperty{}

	p1 := ArtifactoryProperty{
		Name:  "colour",
		Value: "red",
	}
	p2 := ArtifactoryProperty{
		Name:  "model",
		Value: "tesla",
	}
	props = append(props, p1)
	props = append(props, p2)
	response, err := v.Artifacts.UploadFileContents(context.Background(), "example-repo-local", "prova/path/prova.txt", "text/plain", "./fixtures/prova.txt", props)
	assert.Nil(t, err)
	assert.Equal(t, 201, response.StatusCode)

	query := `items.find({ "repo": "example-repo-local", "name": { "$match": "prova.txt" } }).include("name","repo","path","actual_md5","actual_sha1","size","type","property")`
	aqlRes, _, err := v.Artifacts.SearchByAQL(context.Background(), query)
	assert.Nil(t, err)
	assert.Equal(t, "prova.txt", *aqlRes.Results[0].Name)
	assert.Equal(t, "prova/path", *aqlRes.Results[0].Path)
	assert.Equal(t, "example-repo-local", *aqlRes.Results[0].Repo)
	assert.Equal(t, "colour", *aqlRes.Results[0].Properties[1].Key)
	assert.Equal(t, "red", *aqlRes.Results[0].Properties[1].Value)
	assert.Equal(t, "model", *aqlRes.Results[0].Properties[0].Key)
	assert.Equal(t, "tesla", *aqlRes.Results[0].Properties[0].Value)
	filepath := fmt.Sprintf("%sprova.txt", os.TempDir())
	
	ff, err := os.Create(filepath)
	assert.Nil(t, err)
	response, err = v.Artifacts.DownloadFileContents(context.Background(), "example-repo-local", "prova/path/prova.txt", ff)
	assert.Equal(t, 200, response.StatusCode)
}
