package createbootstrapcluster

import (
	"context"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	config2 "github.com/giantswarm/capi-bootstrap/pkg/config"
	"github.com/giantswarm/capi-bootstrap/pkg/kind"
)

func (r *Runner) Run(cmd *cobra.Command, _ []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	bootstrapConfig, err := r.flag.ToConfig()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context(), bootstrapConfig)
	return microerror.Mask(err)
}

func (r *Runner) Do(ctx context.Context, bootstrapConfig config2.BootstrapConfig) error {
	kindClient := kind.Client{
		ClusterName: bootstrapConfig.Spec.BootstrapCluster.Name,
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

		err = os.WriteFile(bootstrapConfig.Spec.BootstrapCluster.Kubeconfig, kubeconfigData, 0644)
		if err != nil {
			return microerror.Mask(err)
		}
	} else {
		kubeconfigData, err = kindClient.CreateCluster(bootstrapConfig.Spec.BootstrapCluster.Kubeconfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	r.logger.Debugf(ctx, "wrote bootstrap cluster kubeconfig to %s", bootstrapConfig.Spec.BootstrapCluster.Kubeconfig)
	r.logger.Debugf(ctx, "created bootstrap cluster")

	return nil
}
