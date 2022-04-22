package pivot

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/pkg/config"
	"github.com/giantswarm/capi-bootstrap/pkg/shell"
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
	r.logger.Debugf(ctx, "pivoting management cluster")

	var source ClusterScope
	{
		var err error
		source.KubeconfigPath, err = environment.GetKubeconfig()
		if err != nil {
			return microerror.Mask(err)
		}
		source.K8sClient, err = environment.GetK8sClient()
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var target ClusterScope
	{
		// TODO: fix this
		var err error
		source.KubeconfigPath, err = environment.GetKubeconfig()
		if err != nil {
			return microerror.Mask(err)
		}
		source.K8sClient, err = environment.GetK8sClient()
		if err != nil {
			return microerror.Mask(err)
		}
	}

	_, stdErr, err := shell.Execute(shell.Command{
		Name: "clusterctl",
		Args: []string{
			"move",
			"--namespace",
			environment.ConfigFile.Spec.ClusterNamespace,
			"--kubeconfig",
			source.KubeconfigPath,
			"--to-kubeconfig",
			target.KubeconfigPath,
		},
	})
	if err != nil {
		return microerror.Mask(fmt.Errorf("%w: %s", err, stdErr))
	}

	err = target.K8sClient.ApplyResources(ctx, environment.ConfigFile.Spec.FileInputs)
	if err != nil {
		return microerror.Mask(err)
	}

	err = target.K8sClient.WaitForClusterReady(ctx, environment.ConfigFile.Spec.ClusterNamespace, environment.ConfigFile.Spec.PermanentCluster.Name)
	if err != nil {
		return microerror.Mask(err)
	}

	err = source.K8sClient.DeleteResources(ctx, environment.ConfigFile.Spec.FileInputs)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "pivoted management cluster")

	return nil
}
