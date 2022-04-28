package generate

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/generator/config"
	"github.com/giantswarm/capi-bootstrap/pkg/lastpass"
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

func (r *Runner) Do(ctx context.Context, _ *cobra.Command, _ []string) error {
	awsSession, err := session.NewSession()
	if err != nil {
		return microerror.Mask(err)
	}

	lastpassClient, err := lastpass.New()
	if err != nil {
		return microerror.Mask(err)
	}

	generators, err := buildGenerators(config.Config{
		AWSSession:     awsSession,
		LastpassClient: lastpassClient,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	templateSecrets, _, err := templates.LoadProvider(r.flag.Provider)
	if err != nil {
		return microerror.Mask(err)
	}

	secrets := map[string]string{}
	for _, secretDefinition := range templateSecrets {
		gen, ok := generators[secretDefinition.Generator]
		if !ok {
			return microerror.Maskf(invalidConfigError, "invalid generator %s", secretDefinition.Generator)
		}

		secretDefinition.BaseDomain = r.flag.BaseDomain
		secretDefinition.ClusterName = r.flag.ClusterName

		generated, err := gen.Generate(ctx, secretDefinition)
		if err != nil {
			return microerror.Mask(err)
		}

		secretYAML, err := yaml.Marshal(generated)
		if err != nil {
			return microerror.Mask(err)
		}

		secrets[secretDefinition.Key] = strings.TrimSpace(string(secretYAML))
	}

	secret := core.Secret{
		TypeMeta: v1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "generated-secrets",
			Namespace: "giantswarm",
		},
		StringData: secrets,
	}
	rendered, err := yaml.Marshal(secret)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = r.stdout.Write(rendered)
	return microerror.Mask(err)
}
