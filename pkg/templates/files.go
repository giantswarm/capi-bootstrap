package templates

import (
	"embed"
	_ "embed"
	"errors"
	"io/fs"
	"strings"

	"github.com/giantswarm/microerror"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/key"
)

//go:embed openstack
var openstackDirectory embed.FS

func LoadProvider(provider string) ([]TemplateSecret, []TemplateFile, error) {
	var directory fs.FS
	switch provider {
	case "openstack":
		var err error
		directory, err = fs.Sub(openstackDirectory, "openstack")
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	default:
		return nil, nil, microerror.Mask(errors.New("invalid provider"))
	}

	secretsFile, err := fs.ReadFile(directory, "secrets.yaml")
	if err != nil {
		return nil, nil, microerror.Mask(err)
	}

	var secrets []TemplateSecret
	err = yaml.Unmarshal(secretsFile, &secrets)
	if err != nil {
		return nil, nil, microerror.Mask(err)
	}

	var templates []TemplateFile
	err = fs.WalkDir(directory, "config", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		content, err := fs.ReadFile(directory, path)
		if err != nil {
			return microerror.Mask(err)
		}
		templates = append(templates, TemplateFile{
			Path:    strings.TrimPrefix(path, "config/"),
			Content: string(content),
		})
		return nil
	})
	if err != nil {
		return nil, nil, microerror.Mask(err)
	}

	err = validateSecrets(secrets)
	if err != nil {
		return nil, nil, microerror.Mask(err)
	}

	err = validateTemplateFiles(templates)
	return secrets, templates, microerror.Mask(err)
}

func validateSecrets(secrets []TemplateSecret) error {
	keys := map[string]struct{}{}
	for i, secretDefinition := range secrets {
		// Make sure key isn't empty
		if secretDefinition.Key == "" {
			return microerror.Maskf(invalidTemplateError, "key for secret with index %d must be defined", i)
		}

		// Make sure generator is defined
		if secretDefinition.Generator == "" {
			return microerror.Maskf(invalidTemplateError, "generator for secret %s must be defined", secretDefinition.Key)
		}

		// Make sure key is unique
		if _, ok := keys[secretDefinition.Key]; ok {
			return microerror.Maskf(invalidTemplateError, "found duplicate key %s", secretDefinition.Key)
		}
		keys[secretDefinition.Key] = struct{}{}

		switch secretDefinition.Generator {
		case key.GeneratorNameAWSIAM:
			return microerror.Maskf(invalidTemplateError, "generator %s is not yet implemented", key.GeneratorNameAWSIAM)
		case key.GeneratorNameCA:
			// fall through
		case key.GeneratorNameGitHubOAuth:
			// fall through
		case key.GeneratorNameLastpass:
			if secretDefinition.Lastpass == nil {
				return microerror.Maskf(invalidTemplateError, "lastpass generator inputs missing for secret %s", secretDefinition.Key)
			}
			if secretDefinition.Lastpass.Format != "yaml" && secretDefinition.Lastpass.Format != "" {
				return microerror.Maskf(invalidTemplateError, "unknown secret format %s for secret %s", secretDefinition.Lastpass.Format, secretDefinition.Key)
			}
			secretRef := secretDefinition.Lastpass.SecretRef
			if secretRef.Name == "" {
				return microerror.Maskf(invalidTemplateError, "lastpass secret name is required for secret %s", secretDefinition.Key)
			}
		case key.GeneratorNameTaylorbot:
			// fall through
		default:
			return microerror.Maskf(invalidTemplateError, "unknown generator %s for secret %s", secretDefinition.Generator, secretDefinition.Key)
		}
	}

	return nil
}

func validateTemplateFiles(templateFiles []TemplateFile) error {
	paths := map[string]struct{}{}
	for i, template := range templateFiles {
		if template.Path == "" {
			return microerror.Maskf(invalidTemplateError, "path is required for template at index %d", i)
		}

		// Make sure path is unique
		if _, ok := paths[template.Path]; ok {
			return microerror.Maskf(invalidTemplateError, "found duplicate path %s", template.Path)
		}
		paths[template.Path] = struct{}{}
	}

	return nil
}
