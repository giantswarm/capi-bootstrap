package createpermanentcluster

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/pkg/config"
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

func (r *Runner) Do(ctx context.Context, environment *config.Environment) error {
	k8sClient, err := environment.GetK8sClient()
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "creating permanent cluster")

	err = k8sClient.CreateNamespace(ctx, environment.ConfigFile.Spec.ClusterNamespace)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		openrcSecret, ok := environment.Secrets["cloud-config"]
		if !ok || openrcSecret == "" {
			return microerror.Maskf(invalidConfigError, "cloud-config secret not found")
		}

		err = k8sClient.CreateCloudConfigSecret(ctx, openrcSecret, environment.ConfigFile.Spec.ClusterNamespace)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = k8sClient.ApplyResources(ctx, environment.ConfigFile.Spec.FileInputs)
	if err != nil {
		return microerror.Mask(err)
	}

	err = k8sClient.WaitForClusterReady(ctx, environment.ConfigFile.Spec.ClusterNamespace, environment.ConfigFile.Spec.PermanentCluster.Name)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "created permanent cluster")

	return nil
}
