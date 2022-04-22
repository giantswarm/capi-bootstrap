package deletepermanentcluster

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
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Runner) Do(ctx context.Context, environment *config.Environment) error {
	k8sClient, err := environment.GetK8sClient()
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "deleting permanent cluster")

	err = k8sClient.DeleteResources(ctx, environment.ConfigFile.Spec.FileInputs)
	if err != nil {
		return microerror.Mask(err)
	}

	err = k8sClient.WaitForClusterDeleted(ctx, environment.ConfigFile.Spec.ClusterNamespace, environment.ConfigFile.Spec.PermanentCluster.Name)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "deleted permanent cluster")

	return nil
}
