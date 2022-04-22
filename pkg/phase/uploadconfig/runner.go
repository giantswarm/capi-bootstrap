package uploadconfig

import (
	"context"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	config2 "github.com/giantswarm/capi-bootstrap/pkg/config"
)

func (r *Runner) Run(cmd *cobra.Command, _ []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	environment, err := r.flag.BuildEnvironment(r.logger)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context(), environment)
	return microerror.Mask(err)
}

func (r *Runner) Do(ctx context.Context, environment *config2.Environment) error {
	k8sClient, err := environment.GetK8sClient()
	if err != nil {
		return microerror.Mask(err)
	}

	lastpassClient, err := environment.GetLastpassClient()
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "uploading config")

	content, err := yaml.Marshal(environment.ConfigFile)
	if err != nil {
		return microerror.Mask(err)
	}

	err = k8sClient.CreateNamespace(ctx, "giantswarm")
	if err != nil {
		return microerror.Mask(err)
	}

	secrets := map[string]string{}
	for _, secret := range environment.ConfigFile.Spec.Secrets {
		var value string
		if secret.EnvVar != nil {
			var ok bool
			value, ok = os.LookupEnv(secret.EnvVar.Name)
			if !ok {
				return microerror.Maskf(notFoundError, "environment variable %s not defined for secret %s", secret.EnvVar.Name, secret.Key)
			}
		} else if secret.Lastpass != nil {
			account, err := lastpassClient.Account(ctx, secret.Lastpass.Share, secret.Lastpass.Group, secret.Lastpass.Name)
			if err != nil {
				return microerror.Mask(err)
			}

			value = account.Notes
		} else {
			return microerror.Maskf(invalidConfigError, "secret definition with key %s is invalid", secret.Key)
		}

		secrets[secret.Key] = value
	}

	resources := []client.Object{
		&core.ConfigMap{
			ObjectMeta: meta.ObjectMeta{
				Name:      "capi-bootstrap",
				Namespace: "giantswarm",
			},
			Data: map[string]string{
				"config": string(content),
			},
		},
		&core.Secret{
			ObjectMeta: meta.ObjectMeta{
				Name:      "capi-bootstrap",
				Namespace: "giantswarm",
			},
			StringData: secrets,
		},
	}

	err = k8sClient.ApplyResources(ctx, resources)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "uploaded config")

	return nil
}
