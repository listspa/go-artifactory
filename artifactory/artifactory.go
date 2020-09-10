package artifactory

import (
	"github.com/listspa/go-artifactory/v2/artifactory/client"
	"github.com/listspa/go-artifactory/v2/artifactory/v1"
	"github.com/listspa/go-artifactory/v2/artifactory/v2"
	 log "github.com/sirupsen/logrus"
	"net/http"
)

// Artifactory is the container for all the api methods
type Artifactory struct {
	V1 *v1.V1
	V2 *v2.V2
}

// NewClient creates a Artifactory from a provided base url for an artifactory instance and a service Artifactory
func NewClient(baseURL string, httpClient *http.Client, loglvl string) (*Artifactory, error) {
	lvl, err := log.ParseLevel(loglvl)
	if err != nil {
		log.Printf(err.Error())
	}
	log.SetLevel(lvl)
	//log.SetReportCaller(true)
	c, err := client.NewClient(baseURL, httpClient)
	if err != nil {
		return nil, err
	}

	rt := &Artifactory{
		V1: v1.NewV1(c),
		V2: v2.NewV2(c),
	}

	return rt, nil
}
