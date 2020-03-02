package v1

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/listspa/go-artifactory/v2/artifactory/client"
	"github.com/listspa/go-artifactory/v2/artifactory/transport"
)


func TestFileInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/storage/arbitrary-repository/path/to/an/existing/artifact", r.RequestURI)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		dummyRes := `{
  "repo" : "arbitrary-repository",
  "path" : "/path/to/an/existing/artifact",
  "created" : "2019-10-22T07:12:08.538+02:00",
  "createdBy" : "jondoe",
  "lastModified" : "2019-10-22T09:38:55.713+02:00",
  "modifiedBy" : "janedoe",
  "lastUpdated" : "2019-10-22T09:38:55.731+02:00",
  "downloadUri" : "http://%s/arbitrary-repository/path/to/an/existing/artifact",
  "mimeType" : "application/zip",
  "size" : "13400",
  "checksums" : {
    "sha1" : "1bc68542d65869e38eece7cfb1b038104ba7a5fb",
    "md5" : "ccb552c5b0714ced4852c8d696da3387",
    "sha256" : "3a4d369251cdd78d616873e1eb7352f83997949969b020397fabc6e2d18801b9"
  },
  "originalChecksums" : {
    "sha1" : "1bc68542d65869e38eece7cfb1b038104ba7a5fb",
    "md5" : "ccb552c5b0714ced4852c8d696da3387",
    "sha256" : "3a4d369251cdd78d616873e1eb7352f83997949969b020397fabc6e2d18801b9"
  },
  "uri" : "%s"
}`

		_, _ = fmt.Fprint(w, fmt.Sprintf(dummyRes, r.Host, r.RequestURI))
	}))

	c, _ := client.NewClient(server.URL, http.DefaultClient)
	v := NewV1(c)

	fileInfo, _, err := v.Artifacts.FileInfo(context.Background(), "arbitrary-repository", "/path/to/an/existing/artifact")
	assert.Nil(t, err)

	assert.Equal(t, "arbitrary-repository", *fileInfo.Repo)
	assert.Equal(t, "/path/to/an/existing/artifact", *fileInfo.Path)
	assert.Equal(t, "2019-10-22T07:12:08.538+02:00", *fileInfo.Created)
	assert.Equal(t, "jondoe", *fileInfo.CreatedBy)
	assert.Equal(t, "2019-10-22T09:38:55.713+02:00", *fileInfo.LastModified)
	assert.Equal(t, "janedoe", *fileInfo.ModifiedBy)
	assert.Equal(t, "2019-10-22T09:38:55.731+02:00", *fileInfo.LastUpdated)
	assert.Equal(t, fmt.Sprintf("%s/arbitrary-repository/path/to/an/existing/artifact", server.URL), *fileInfo.DownloadUri)
	assert.Equal(t, "application/zip", *fileInfo.MimeType)
	assert.Equal(t, 13400, *fileInfo.Size)
	assert.Equal(t, "1bc68542d65869e38eece7cfb1b038104ba7a5fb", *fileInfo.Checksums.Sha1)
	assert.Equal(t, "1bc68542d65869e38eece7cfb1b038104ba7a5fb", *fileInfo.OriginalChecksums.Sha1)
	assert.Equal(t, "ccb552c5b0714ced4852c8d696da3387", *fileInfo.Checksums.Md5)
	assert.Equal(t, "ccb552c5b0714ced4852c8d696da3387", *fileInfo.OriginalChecksums.Md5)
	assert.Equal(t, "3a4d369251cdd78d616873e1eb7352f83997949969b020397fabc6e2d18801b9", *fileInfo.Checksums.Sha256)
	assert.Equal(t, "3a4d369251cdd78d616873e1eb7352f83997949969b020397fabc6e2d18801b9", *fileInfo.OriginalChecksums.Sha256)
	assert.Equal(t, "/api/storage/arbitrary-repository/path/to/an/existing/artifact", *fileInfo.Uri)
}

func TestSearchFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/search/aql", r.RequestURI)
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		assert.Equal(t, `items.find({ "repo": "clibs-local", "name": { "$match": "RdbManager-2.1.13d4-plat-*.zip" } }).include("name","repo","path","actual_md5","actual_sha1","size","type","property")`, bodyString)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		dummyRes := `{
			"results": [
			  {
				"repo": "clibs-local",
				"path": "cm/dev/libs/lift/rdb2/RdbManager/2.1.13d4",
				"name": "RdbManager-2.1.13d4-plat-NT9_64.zip",
				"type": "file",
				"size": 84600,
				"actual_md5": "f3b8c256afc3bad62d6c7da02a39c785",
				"actual_sha1": "18854d9d279ff0784e645accec49b3a5fc5194bd",
				"properties": [
				  {
					"key": "type",
					"value": "NT9_64"
				  }
				]
			  },
			  {
				"repo": "clibs-local",
				"path": "cm/dev/libs/lift/rdb2/RdbManager/2.1.13d4",
				"name": "RdbManager-2.1.13d4-plat-NT9_32.zip",
				"type": "file",
				"size": 79943,
				"actual_md5": "593573c4e9b50c7957fa2e0fabf17c20",
				"actual_sha1": "21088188a84faec774b63bc511b4390d357a12fc",
				"properties": [
				  {
					"key": "type",
					"value": "NT9_32"
				  }
				]
			  },
			  {
				"repo": "clibs-local",
				"path": "cm/dev/libs/lift/rdb2/RdbManager/2.1.13d4",
				"name": "RdbManager-2.1.13d4-plat-LX5_32.zip",
				"type": "file",
				"size": 34451,
				"actual_md5": "fa9dc6a25be9de63916f1e629a2444a4",
				"actual_sha1": "5e67b4e27274b952a2a6686aca94e3860d8c9e45",
				"properties": [
				  {
					"key": "type",
					"value": "LX5_32"
				  }
				]
			  },
			  {
				"repo": "clibs-local",
				"path": "cm/dev/libs/lift/rdb2/RdbManager/2.1.13d4",
				"name": "RdbManager-2.1.13d4-plat-LX7_64.zip",
				"type": "file",
				"size": 39710,
				"actual_md5": "d6eb3fcf48faad002e4893094d76f08f",
				"actual_sha1": "7d25fa739c39f9a6ab3eb6e657e0e458f41b9eed",
				"properties": [
				  {
					"key": "type",
					"value": "LX7_64"
				  }
				]
			  },
			  {
				"repo": "clibs-local",
				"path": "cm/dev/libs/lift/rdb2/RdbManager/2.1.13d4",
				"name": "RdbManager-2.1.13d4-plat-AX72_64.zip",
				"type": "file",
				"size": 41129,
				"actual_md5": "7e697e4d820f6756855ec226beb026be",
				"actual_sha1": "3a16c925faf9964247a79c00bca722572c388ed5",
				"properties": [
				  {
					"key": "type",
					"value": "AX72_64"
				  }
				]
			  },
			  {
				"repo": "clibs-local",
				"path": "cm/dev/libs/lift/rdb2/RdbManager/2.1.13d4",
				"name": "RdbManager-2.1.13d4-plat-LX5_64.zip",
				"type": "file",
				"size": 37330,
				"actual_md5": "311ad274442e289f320828df3f36caa1",
				"actual_sha1": "fef57afa72c0eabe2ac2f92b009c8defda4d2838",
				"properties": [
				  {
					"key": "type",
					"value": "LX5_64"
				  }
				]
			  }
			],
			"range": {
			  "start_pos": 0,
			  "end_pos": 6,
			  "total": 6
			}
		  }`
		_, _ = fmt.Fprint(w, fmt.Sprintf(dummyRes, r.Host, r.RequestURI))
	}))
	c, _ := client.NewClient(server.URL, http.DefaultClient)
	v := NewV1(c)
    query := `items.find({ "repo": "clibs-local", "name": { "$match": "RdbManager-2.1.13d4-plat-*.zip" } }).include("name","repo","path","actual_md5","actual_sha1","size","type","property")`
	results, _, err := v.Artifacts.SearchByAQL(context.Background(), query)
	assert.Nil(t, err)
	assert.NotNil(t, results)
	assert.NotEmpty(t, results)
	assert.Equal(t, 6, len(results.Results))

	assert.Equal(t, "RdbManager-2.1.13d4-plat-NT9_64.zip", *results.Results[0].Name)
	assert.Equal(t, 84600, *results.Results[0].Size)
	assert.Equal(t, "cm/dev/libs/lift/rdb2/RdbManager/2.1.13d4", *results.Results[0].Path)
	assert.Equal(t, "type", *results.Results[0].Properties[0].Key)
	assert.Equal(t, "NT9_64", *results.Results[0].Properties[0].Value)

	assert.Equal(t, "RdbManager-2.1.13d4-plat-NT9_32.zip", *results.Results[1].Name)
	assert.Equal(t, 79943, *results.Results[1].Size)
	assert.Equal(t, "cm/dev/libs/lift/rdb2/RdbManager/2.1.13d4", *results.Results[1].Path)
	assert.Equal(t, "type", *results.Results[1].Properties[0].Key)
	assert.Equal(t, "NT9_32", *results.Results[1].Properties[0].Value)
}

func TestSearchFilesNoResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/search/aql", r.RequestURI)
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		assert.Equal(t, `items.find({ "repo": "clibs-local", "name": { "$match": "RdbManager-2.1.13d4-plat-*.zip" } }).include("name","repo","path","actual_md5","actual_sha1","size","type","property")`, bodyString)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		dummyRes := `{
			"results": [],
			"range": {
			  "start_pos": 0,
			  "end_pos": 0,
			  "total": 0
			}
		  }`
		_, _ = fmt.Fprint(w, fmt.Sprintf(dummyRes, r.Host, r.RequestURI))
	}))
	c, _ := client.NewClient(server.URL, http.DefaultClient)
	v := NewV1(c)
    query := `items.find({ "repo": "clibs-local", "name": { "$match": "RdbManager-2.1.13d4-plat-*.zip" } }).include("name","repo","path","actual_md5","actual_sha1","size","type","property")`
	results, _, err := v.Artifacts.SearchByAQL(context.Background(), query)
	assert.Nil(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results.Results))
	assert.Nil(t, err)
}

func TestSearchFilesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/search/aql", r.RequestURI)
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		assert.Equal(t, `items.find({ "repo": "clibs-local", "name": { "$match": "RdbManager-2.1.13d4-plat-*.zip" } }).include("name","repo","path","actual_md5","actual_sha1","size","type","property")`, bodyString)
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		dummyRes := `{"errors": [{"status": 404,"message": "Not Found"}]}`
		_, _ = fmt.Fprint(w, fmt.Sprintf(dummyRes, r.Host, r.RequestURI))
	}))
	c, _ := client.NewClient(server.URL, http.DefaultClient)
	v := NewV1(c)
	query := `items.find({ "repo": "clibs-local", "name": { "$match": "RdbManager-2.1.13d4-plat-*.zip" } }).include("name","repo","path","actual_md5","actual_sha1","size","type","property")`
	results, response, err := v.Artifacts.SearchByAQL(context.Background(), query)
	assert.NotNil(t, err)
	assert.Nil(t, results.Results)
	assert.Equal(t, 404, response.StatusCode)
	assert.Equal(t, "404 Not Found", response.Status)
}

func TestDownloadFileContents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//check wellformed request
		assert.Equal(t, "/arbitrary-repository/path/to/an/existing/artifact", r.RequestURI)
		assert.Equal(t, "GET", r.Method)
		usr, pwd, ok := r.BasicAuth()
		assert.Equal(t, "admin", usr)
		assert.Equal(t, "password", pwd)
		assert.True(t, ok)
		authH := r.Header.Get("Authorization")
		assert.Equal(t, "Basic YWRtaW46cGFzc3dvcmQ=", authH)

		//response
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		res := "dummy content"
		_, _ = fmt.Fprint(w, res)
	}))
	tp := transport.BasicAuth{
		Username: "admin",
		Password: "password",
	}
	c, _ := client.NewClient(server.URL, tp.Client())
	v := NewV1(c)

	target := bytes.NewBufferString("")
	response, err := v.Artifacts.DownloadFileContents(context.Background(), "arbitrary-repository", "path/to/an/existing/artifact", target)

	assert.Equal(t, "dummy content", target.String())
	assert.NotNil(t, 200, response.StatusCode)
	assert.Nil(t, err)
}

func TestUploadFileContents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//check wellformed request
		assert.Equal(t, "/clibs-local/prova/path/prova.txt", r.RequestURI)
		assert.Equal(t, "PUT", r.Method)
		usr, pwd, ok := r.BasicAuth()
		assert.Equal(t, "admin", usr)
		assert.Equal(t, "password", pwd)
		assert.True(t, ok)
		authH := r.Header.Get("Authorization")
		assert.Equal(t, "Basic YWRtaW46cGFzc3dvcmQ=", authH)
		contentH := r.Header.Get("Content-Type")
		assert.Equal(t, "text/plain", contentH)
		//response
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		dummyRes := `{}`
		_, _ = fmt.Fprint(w, fmt.Sprintf(dummyRes, r.Host, r.RequestURI))
	}))
	tp := transport.BasicAuth{
		Username: "admin",
		Password: "password",
	}
	c, _ := client.NewClient(server.URL, tp.Client())
	v := NewV1(c)

	response, err := v.Artifacts.UploadFileContents(context.Background(), "clibs-local", "prova/path/prova.txt", "text/plain", "./fixtures/prova.txt", []ArtifactoryProperty{})
	assert.Nil(t, err)
	assert.Equal(t, 201, response.StatusCode)

}

func TestUploadFileContentsWithProperties(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//check wellformed request
		assert.Equal(t, "/clibs-local/prova/path/prova.txt;type=text;color=red", r.RequestURI)
		assert.Equal(t, "PUT", r.Method)
		usr, pwd, ok := r.BasicAuth()
		assert.Equal(t, "admin", usr)
		assert.Equal(t, "password", pwd)
		assert.True(t, ok)
		authH := r.Header.Get("Authorization")
		assert.Equal(t, "Basic YWRtaW46cGFzc3dvcmQ=", authH)
		contentH := r.Header.Get("Content-Type")
		assert.Equal(t, "text/plain", contentH)
		md5sum := r.Header.Get("X-Checksum-MD5")
		assert.Equal(t, "05284e7b404c38ff2e298f75268cfd49", md5sum)

		//response
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		dummyRes := `{}`
		_, _ = fmt.Fprint(w, fmt.Sprintf(dummyRes, r.Host, r.RequestURI))
	}))
	tp := transport.BasicAuth{
		Username: "admin",
		Password: "password",
	}
	c, _ := client.NewClient(server.URL, tp.Client())
	v := NewV1(c)

	propt1 := ArtifactoryProperty{
		Name:  "type",
		Value: "text",
	}
	propt2 := ArtifactoryProperty{
		Name:  "color",
		Value: "red",
	}
	response, err := v.Artifacts.UploadFileContents(context.Background(), "clibs-local", "prova/path/prova.txt", "text/plain", "./fixtures/prova.txt", []ArtifactoryProperty{propt1, propt2})
	assert.Nil(t, err)
	assert.Equal(t, 201, response.StatusCode)

}
