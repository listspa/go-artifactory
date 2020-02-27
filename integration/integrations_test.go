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
var repo = os.Getenv("DEFAULT_REPO")
func TestSearchFiles(t *testing.T) {
	tp := transport.BasicAuth{
		Username: user,
		Password: pwd,
	}
	client, err := artifactory.NewClient(url, tp.Client())
	assert.Nil(t, err)
	results, _, err := client.V1.Artifacts.SearchFiles(context.Background(), repo, "RdbManager-2.1.13d4-plat-*.zip")
	assert.Nil(t, err)
	assert.Equal(t, 6, len(results.Results))
}

func TestSearchAndDownload(t *testing.T) {
	tp := transport.BasicAuth{
		Username: user,
		Password: pwd,
	}
	client, err := artifactory.NewClient(url, tp.Client())
	assert.Nil(t, err)
	results, _, err := client.V1.Artifacts.SearchFiles(context.Background(), repo, "RdbManager-2.1.13d4-plat-*.zip")
	assert.Nil(t, err)
	assert.Equal(t, 6, len(results.Results))

	for _, rr := range results.Results {
		f, err := os.Create(os.TempDir()+"/"+*rr.Name)
		assert.Nil(t, err)
		downloadPath := fmt.Sprintf("%s/%s", *rr.Path, *rr.Name)
		finfo, _, err := client.V1.Artifacts.DownloadFileContents(context.Background(), repo, downloadPath, f)
		assert.Nil(t, err)
		assert.NotNil(t, finfo)
		statinfo, err := f.Stat()
		assert.Nil(t, err)
		assert.False(t, statinfo.IsDir())
		assert.False(t, statinfo.Size() == 0)
		defer f.Close()
	}

}
func TestUpload(t *testing.T) {
	tp := transport.BasicAuth{
		Username: user,
		Password: pwd,
	}
	client, err := artifactory.NewClient(url, tp.Client())
	assert.Nil(t, err)
	file,err := os.Open("prova.txt")
	assert.Nil(t, err)
	finfo, _, err := client.V1.Artifacts.UploadFileContents(context.Background(), repo, "prova/path/prova.txt", file)
	assert.Nil(t, err)
	assert.NotNil(t, finfo)
	assert.Equal(t, *finfo.Repo, repo)
	assert.Equal(t, *finfo.Path, "/prova/path/prova.txt")
}
