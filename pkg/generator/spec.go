package generator

import (
	"context"

	"github.com/giantswarm/capi-bootstrap/pkg/templates"
)

type Config struct {
}

type Generator interface {
	Generate(ctx context.Context, templateSecret templates.TemplateSecret, installationInputs templates.InstallationInputs) (interface{}, error)
}
