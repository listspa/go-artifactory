package v1

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

var searchTemplate = `items.find({"repo": "%s","path": {"$ne": "."},"$or": [{"$and":[{"path": {"$match": "*"},"name": {"$match": "%s"}}]}]}).include("name","repo","path","actual_md5","actual_sha1","size","type","property")`

func init() {
	log.SetPrefix("[Artifactory-Client] ")
}

// ArtifactService exposes the Artifact API endpoints from Artifactory
type ArtifactService Service

// SingleReplicationConfig is the model of the Artifactory Replication Config
type SingleReplicationConfig struct {
	RepoKey                *string `json:"repoKey,omitempty"`
	URL                    *string `json:"url,omitempty"`
	SocketTimeoutMillis    *int    `json:"socketTimeoutMillis,omitempty"`
	Username               *string `json:"username,omitempty"`
	Password               *string `json:"password,omitempty"`
	Enabled                *bool   `json:"enabled,omitempty"`
	SyncDeletes            *bool   `json:"syncDeletes,omitempty"`
	SyncProperties         *bool   `json:"syncProperties,omitempty"`
	SyncStatistics         *bool   `json:"syncStatistics,omitempty"`
	PathPrefix             *string `json:"pathPrefix,omitempty"`
	CronExp                *string `json:"cronExp,omitempty"` // Only required when getting list of repositories as C*UD operations will be done through a repConfig obj
	EnableEventReplication *bool   `json:"enableEventReplication,omitempty"`
}

// ReplicationConfig is the model for the multi replication config API endpoints. Its usage is preferred over
// SingleReplicationConfig as it is a more direct mapping of the replicationConfig in the UI
type ReplicationConfig struct {
	RepoKey                *string                    `json:"-"`
	CronExp                *string                    `json:"cronExp,omitempty"`
	EnableEventReplication *bool                      `json:"enableEventReplication,omitempty"`
	Replications           *[]SingleReplicationConfig `json:"replications,omitempty"`
}

func (r ReplicationConfig) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Creates or replaces a local multi-push replication configuration. Supported by local repositories.
// Notes: Requires an enterprise license
// Security: Requires a privileged user
func (s *ArtifactService) SetRepositoryReplicationConfig(ctx context.Context, repoKey string, config *ReplicationConfig) (*http.Response, error) {
	path := fmt.Sprintf("/api/replications/multiple/%s", repoKey)
	req, err := s.client.NewJSONEncodedRequest("PUT", path, config)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Description: Add or replace replication configuration for given repository key. Supported by local and remote repositories. Accepts the JSON payload returned from Get Repository Replication Configuration for a single and an array of configurations. If the payload is an array of replication configurations, then values for cronExp and enableEventReplication in the first element in the array will determine the corresponding values when setting the repository replication configuration.
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *ArtifactService) SetSingleRepositoryReplicationConfig(ctx context.Context, repoKey string, config *SingleReplicationConfig) (*http.Response, error) {
	path := fmt.Sprintf("/api/replications/%s", repoKey)
	req, err := s.client.NewJSONEncodedRequest("PUT", path, config)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Returns the replication configuration for the given repository key, if found. Supported by local and remote repositories. Note: The 'enableEventReplication' parameter refers to both push and pull replication.
// Notes: Requires Artifactory Pro
// Security: Requires a privileged user
func (s *ArtifactService) GetRepositoryReplicationConfig(ctx context.Context, repoKey string) (*ReplicationConfig, *http.Response, error) {
	replications, resp, err := s.getReplicationConfigs(ctx, repoKey)
	if err != nil {
		return nil, resp, err
	}

	replicationConfig := new(ReplicationConfig)

	if len(replications) > 0 {
		replicationConfig.Replications = new([]SingleReplicationConfig)
	}

	for _, replication := range replications {
		replicationConfig.RepoKey = replication.RepoKey
		replicationConfig.CronExp = replication.CronExp
		replicationConfig.EnableEventReplication = replication.EnableEventReplication

		*replicationConfig.Replications = append(*replicationConfig.Replications, replication)
	}

	return replicationConfig, resp, nil
}

// Gets the replication configs for a given repository key.
// Note: As the get endpoint can return a single object or an array (if there is more than one replication), extra work
// is required to marshall the response into an expected, consistent format.
func (s *ArtifactService) getReplicationConfigs(ctx context.Context, repoKey string) ([]SingleReplicationConfig, *http.Response, error) {
	path := fmt.Sprintf("/api/replications/%s", repoKey)
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeReplicationConfig)

	// By writing the response to a buffer, we can evaluate the type and decode appropriately.
	var httpBody bytes.Buffer
	resp, err := s.client.Do(ctx, req, &httpBody)
	if err != nil {
		return nil, resp, err
	}

	// A copy is required as the initial write to the httpBody buffer contains EOF issues.
	var httpBodyCopy bytes.Buffer
	_, err = io.Copy(&httpBodyCopy, &httpBody)
	if err != nil {
		return nil, resp, err
	}

	var repConfigAsInterface interface{}
	err = json.Unmarshal(httpBodyCopy.Bytes(), &repConfigAsInterface)
	if err != nil {
		return nil, resp, err
	}

	// Checks to see what type of response is returned, and then casts to that.
	replications := make([]SingleReplicationConfig, 0)
	switch repConfigAsInterface.(type) {
	case []interface{}:
		err = json.NewDecoder(&httpBodyCopy).Decode(&replications)
		if err != nil {
			return nil, resp, err
		}
	default:
		singleReplication := new(SingleReplicationConfig)
		err = json.NewDecoder(&httpBodyCopy).Decode(singleReplication)
		if err != nil {
			return nil, resp, err
		}

		replications = append(replications, *singleReplication)
	}

	return replications, resp, nil
}

// Updates a local multi-push replication configuration. Supported by local repositories.
// Notes: Requires an enterprise license
// Security: Requires a privileged user
func (s *ArtifactService) UpdateRepositoryReplicationConfig(ctx context.Context, repoKey string, config *ReplicationConfig) (*http.Response, error) {
	path := fmt.Sprintf("/api/replications/multiple/%s", repoKey)
	req, err := s.client.NewJSONEncodedRequest("POST", path, config)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Update existing replication configuration for given repository key, if found. Supported by local and remote repositories.
// Notes: Requires Artifactory Pro
// Security: Requires a privileged user
func (s *ArtifactService) UpdateSingleRepositoryReplicationConfig(ctx context.Context, repoKey string, config *SingleReplicationConfig) (*http.Response, error) {
	path := fmt.Sprintf("/api/replications/%s", repoKey)
	req, err := s.client.NewJSONEncodedRequest("POST", path, config)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Description: Delete existing replication configuration for given repository key. Supported by local and local-cached repositories.
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *ArtifactService) DeleteRepositoryReplicationConfig(ctx context.Context, repoKey string) (*http.Response, error) {

	path := fmt.Sprintf("/api/replications/%s", repoKey)
	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

type Checksums struct {
	Md5    *string `json:"md5,omitempty"`
	Sha1   *string `json:"sha1,omitempty"`
	Sha256 *string `json:"sha256,omitempty"`
}

type FileInfo struct {
	Repo              *string    `json:"repo,omitempty"`
	Path              *string    `json:"path,omitempty"`
	Created           *string    `json:"created,omitempty"`
	CreatedBy         *string    `json:"createdBy,omitempty"`
	LastModified      *string    `json:"lastModified,omitempty"`
	ModifiedBy        *string    `json:"modifiedBy,omitempty"`
	LastUpdated       *string    `json:"lastUpdated,omitempty"`
	DownloadUri       *string    `json:"downloadUri,omitempty"`
	MimeType          *string    `json:"mimeType,omitempty"`
	Size              *int       `json:"size,string,omitempty"`
	Checksums         *Checksums `json:"checksums,omitempty"`
	OriginalChecksums *Checksums `json:"originalChecksums,omitempty"`
	Uri               *string    `json:"uri,omitempty"`
}

type AqlSearchResults struct {
	Results []AqlResult `json:"results,omitempty"`
}

type AqlResult struct {
	Repo       *string       `json:"repo,omitempty"`
	Path       *string       `json:"path,omitempty"`
	Name       *string       `json:"name,omitempty"`
	Size       *int          `json:"size,omitempty"`
	Properties []AqlProperty `json:"properties,omitempty"`
}

type AqlProperty struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`
}

type ArtifactoryProperty struct {
	Name  string
	Value string
}

// FileInfo Returns the metadata of the given file. Supported by local, local-cached and virtual repositories.
// Security: Requires a privileged user (can be anonymous)
func (s *ArtifactService) FileInfo(ctx context.Context, repoKey string, filePath string) (*FileInfo, *http.Response, error) {
	path := fmt.Sprintf("/api/storage/%s/%s", repoKey, filePath)
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating new request: %v", err)
	}
	req.Header.Set("Accept", mediaTypeFileInfo)
	log.Printf("Storage API [%s]", req.URL.String())
	fileInfo := new(FileInfo)
	resp, err := s.client.Do(ctx, req, fileInfo)
	return fileInfo, resp, err
}

// DownloadFileContents Copies the specified file to the given target. Supported by local, local-cached and virtual repositories.
// Security: Requires a privileged user (can be anonymous)
func (s *ArtifactService) DownloadFileContents(ctx context.Context, repoKey string, filePath string, file io.Writer) (*http.Response, error) {
	if file == nil {
		return nil, fmt.Errorf("target is not allowed to be nil")
	}

	targetURL := fmt.Sprintf("%s%s/%s", s.client.BaseURL.String(), repoKey, filePath)
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating new request: %v", err)
	}
	log.Printf("Downloading API [%s]", req.URL.String())
	resp, err := s.client.Do(ctx, req, file)
	return resp, err

}

// UploadFileContents Copies the specified file to the given target in Artifactory
func (s *ArtifactService) UploadFileContents(ctx context.Context, repoKey string, filePath string, mimetype string, localfile string, props []ArtifactoryProperty) (*http.Response, error) {
	var err error
	var req *http.Request
	
	targetURL := fmt.Sprintf("%s%s/%s", s.client.BaseURL.String(), repoKey, filePath)
	for _, p := range props {
		targetURL = fmt.Sprintf("%s;%s=%s", targetURL, p.Name, p.Value)
	}
	//content
	content, err := ioutil.ReadFile(localfile)
	if err != nil {
		return nil, errors.Wrapf(err, "reading file content [%s]", filePath)
	}

	//md5sum
	file, err := os.Open(localfile)
	if err != nil {
		return nil, errors.Wrapf(err, "opening file [%s]", filePath)
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, errors.Wrapf(err, "reading file m5dsum [%s]", filePath)
	}
	hashInBytes := hash.Sum(nil)[:16]
	
	req, err = http.NewRequest("PUT", targetURL, bytes.NewBuffer(content))
	if err != nil {
		return nil, fmt.Errorf("creating new request: %v", err)
	}
	req.Header.Set("Content-Type", mimetype)
	req.Header.Set("X-Checksum-MD5", hex.EncodeToString(hashInBytes))
	log.Printf("Uploading API [%s]", req.URL.String())
	resp, err := s.client.Do(ctx, req, nil)
	return resp, err

}

// SearchFiles search files using AQL language
func (s *ArtifactService) SearchFiles(ctx context.Context, repoKey string, pattern string) (*AqlSearchResults, *http.Response, error) {
	path := "/api/search/aql"
	query := fmt.Sprintf(searchTemplate, repoKey, pattern)

	body := []byte(query)
	req, err := s.client.NewRequest("POST", path, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, fmt.Errorf("creating new request: %v", err)
	}
	log.Printf("AQL API [%s] query [%s]", req.URL.String(), query)
	aqlresults := new(AqlSearchResults)
	resp, err := s.client.Do(ctx, req, aqlresults)

	return aqlresults, resp, err
}
