package config

import (
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/giantswarm/capi-bootstrap/pkg/lastpass"
)

type Config struct {
	AWSSession     *session.Session
	LastpassClient *lastpass.Client
}
