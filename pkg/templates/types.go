package templates

import (
	"github.com/giantswarm/capi-bootstrap/pkg/generator/secret"
)

type TemplateData struct {
	BaseDomain  string
	ClusterName string
	Customer    string
	Pipeline    string
	Provider    string
	Secrets     map[string]interface{}
}

type TemplateFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type ProviderDefinition struct {
	Secrets   []secret.GeneratedSecretDefinition
	Templates []TemplateFile
}
