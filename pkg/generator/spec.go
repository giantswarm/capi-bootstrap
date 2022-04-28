package generator

import (
	"context"

	"github.com/giantswarm/capi-bootstrap/pkg/generator/secret"
)

type Config struct {
}

type Generator interface {
	Generate(ctx context.Context, templateInputs secret.GeneratedSecretDefinition) (interface{}, error)
}
