package integration

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/susannamartinelli/go-artifactory/v2/artifactory"
	"github.com/susannamartinelli/go-artifactory/v2/artifactory/transport"
)
// "https://artifactory.list-group.com/artifactory"
var url = os.Getenv("ARTIFACTORY_URL")
var user = os.Getenv("ARTIFACTORY_USER")
var pwd = os.Getenv("ARTIFACTORY_PASSWD")

func TestIntegrationSearchFiles(t *testing.T) {
	tp := transport.BasicAuth{
		Username: user,
		Password: pwd,
	}
	

	client, err := artifactory.NewClient(url, tp.Client())

	results, _, err := client.V1.Artifacts.SearchFiles(context.Background(), "RdbManager-2.1.13d4-plat-*")
	assert.Nil(t, err)
	assert.Equal(t, 6, len(results.Results))
	
}
