package v2

import (
	"github.com/susannamartinelli/go-artifactory/v2/artifactory/client"
)

type Service struct {
	client *client.Client
}

type V2 struct {
	common Service

	Security *SecurityService
}
