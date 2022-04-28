package generate

import (
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	texttemplate "text/template"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/generator/secret"
	"github.com/giantswarm/capi-bootstrap/pkg/templates"
)

func (r *Runner) Run(cmd *cobra.Command, args []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context(), cmd, args)
	return microerror.Mask(err)
}

func yamlIsEncrypted(data []byte) (bool, error) {
	var dataAsMap map[string]interface{}
	err := yaml.Unmarshal(data, &dataAsMap)
	if err != nil {
		return false, microerror.Mask(err)
	}
	_, keyExists := dataAsMap["sops"]
	return keyExists, nil
}

func loadInstallationSecrets(path string, templateSecrets []secret.GeneratedSecretDefinition) (map[string]interface{}, error) {
	installationSecretsFile, err := os.ReadFile(path)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if isEncrypted, err := yamlIsEncrypted(installationSecretsFile); err != nil {
		return nil, microerror.Mask(err)
	} else if isEncrypted {
		// TODO: we could decrypt here in the future
		return nil, microerror.Maskf(invalidConfigError, "installation secrets must be decrypted before being passed to this command")
	}

	var asSecret core.Secret
	err = yaml.Unmarshal(installationSecretsFile, &asSecret)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Support both stringData and data (stringData is easier to read since values are not base64 encoded)
	installationSecrets := map[string]interface{}{}
	if asSecret.StringData != nil {
		for key, valueString := range asSecret.StringData {
			value, err := parseValue([]byte(valueString), false)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			installationSecrets[key] = value
		}
	} else if asSecret.Data != nil {
		for key, valueEncoded := range asSecret.Data {
			value, err := parseValue(valueEncoded, true)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			installationSecrets[key] = value
		}
	}

	// Validate installation secrets against expected secrets
	if len(installationSecrets) != len(templateSecrets) {
		return nil, microerror.Maskf(invalidSecretError, "number of secrets didn't match, found %d, expected %d", len(installationSecrets), len(templateSecrets))
	}

	for _, templateSecret := range templateSecrets {
		_, ok := installationSecrets[templateSecret.Key]
		if !ok {
			return nil, microerror.Maskf(invalidSecretError, "didn't find installation secret with key %s", templateSecret.Key)
		}
	}

	return installationSecrets, nil
}

func templateFileToTemplate(file templates.TemplateFile) (*texttemplate.Template, error) {
	funcs := map[string]interface{}{
		"nindent": nindent, // Inject this helper function for dealing with multiline strings (adapted from https://github.com/Masterminds/sprig)
	}
	templateFileTemplate, err := texttemplate.New(file.Path).Funcs(funcs).Parse(file.Content)
	return templateFileTemplate, microerror.Mask(err)
}

func (r *Runner) Do(ctx context.Context, _ *cobra.Command, _ []string) error {
	templateSecrets, templateFiles, err := templates.LoadProvider(r.flag.Provider)
	if err != nil {
		return microerror.Mask(err)
	}

	installationSecrets, err := loadInstallationSecrets(r.flag.InstallationSecretsFile, templateSecrets)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, file := range templateFiles {
		template, err := templateFileToTemplate(file)
		if err != nil {
			return microerror.Mask(err)
		}

		var rendered bytes.Buffer
		err = template.Execute(&rendered, templates.TemplateData{
			BaseDomain:  r.flag.BaseDomain,
			ClusterName: r.flag.ClusterName,
			Customer:    r.flag.Customer,
			Pipeline:    r.flag.Pipeline,
			Provider:    r.flag.Provider,
			Secrets:     installationSecrets,
		})
		if err != nil {
			return microerror.Mask(err)
		}

		fileDirectory := filepath.Join(r.flag.OutputDirectory, filepath.Dir(file.Path))
		err = os.MkdirAll(fileDirectory, 0755)
		if err != nil {
			return microerror.Mask(err)
		}

		filePath := filepath.Join(r.flag.OutputDirectory, file.Path)
		err = os.WriteFile(filePath, rendered.Bytes(), 0644)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return pad + strings.Replace(v, "\n", "\n"+pad, -1)
}

func nindent(spaces int, v string) string {
	return "\n" + indent(spaces, v)
}

func parseValue(valueBytes []byte, base64Encoded bool) (interface{}, error) {
	valueDecoded := valueBytes

	if base64Encoded {
		var err error
		valueDecoded, err = base64.StdEncoding.DecodeString(string(valueBytes))
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var valueParsed interface{}
	err := yaml.Unmarshal(valueDecoded, &valueParsed)
	return valueParsed, microerror.Mask(err)
}
