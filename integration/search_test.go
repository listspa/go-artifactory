package integration

import (
	"context"
	"fmt"
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
	results, _, err := client.V1.Artifacts.SearchFiles(context.Background(), "clibs-local", "RdbManager-2.1.13d4-plat-*.zip")
	assert.Nil(t, err)
	assert.Equal(t, 6, len(results.Results))
}

func TestIntegrationSearchAndDownload(t *testing.T) {
	tp := transport.BasicAuth{
		Username: user,
		Password: pwd,
	}
	client, err := artifactory.NewClient(url, tp.Client())
	results, _, err := client.V1.Artifacts.SearchFiles(context.Background(), "clibs-local", "RdbManager-2.1.13d4-plat-*.zip")
	assert.Nil(t, err)
	assert.Equal(t, 6, len(results.Results))

	for _, rr := range results.Results {
		f, err := os.Create(os.TempDir()+"/"+*rr.Name)
		assert.Nil(t, err)
		downloadPath := fmt.Sprintf("%s/%s", *rr.Path, *rr.Name)
		client.V1.Artifacts.FileContents(context.Background(), "clibs-local", downloadPath, f)
		statinfo, err := f.Stat()
		assert.Nil(t, err)
		assert.False(t, statinfo.IsDir())
		assert.False(t, statinfo.Size() == 0)
		defer f.Close()
	}

}
