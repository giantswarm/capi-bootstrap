package createbootstrapcluster

import (
	"context"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

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
	kindClient, err := environment.GetKindClient()
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "creating bootstrap cluster")

	var kubeconfigData []byte
	if exists, err := kindClient.ClusterExists(); err != nil {
		return microerror.Mask(err)
	} else if exists {
		r.logger.Debugf(ctx, "bootstrap cluster already exists")

		kubeconfigData, err = kindClient.GetKubeconfig()
		if err != nil {
			return microerror.Mask(err)
		}

		err = os.WriteFile(environment.ConfigFile.Spec.BootstrapCluster.Kubeconfig, kubeconfigData, 0644)
		if err != nil {
			return microerror.Mask(err)
		}
	} else {
		kubeconfigData, err = kindClient.CreateCluster(environment.ConfigFile.Spec.BootstrapCluster.Kubeconfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	r.logger.Debugf(ctx, "wrote bootstrap cluster kubeconfig to %s", environment.ConfigFile.Spec.BootstrapCluster.Kubeconfig)
	r.logger.Debugf(ctx, "created bootstrap cluster")

	return nil
}
