package taylorbot

import (
	"github.com/giantswarm/capi-bootstrap/pkg/lastpass"
)

type Generator struct {
	client *lastpass.Client
}
