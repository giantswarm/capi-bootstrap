package deletebootstrapcluster

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/pkg/kind"
)

func (r *Runner) Run(cmd *cobra.Command, args []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context(), cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Runner) Do(ctx context.Context, _ *cobra.Command, _ []string) error {
	kindClient := kind.Client{
		ClusterName: r.flag.KindClusterName,
	}

	r.logger.Debugf(ctx, "creating bootstrap cluster")

	if exists, err := kindClient.ClusterExists(); err != nil {
		return microerror.Mask(err)
	} else if exists {
		err = kindClient.DeleteCluster()
		if err != nil {
			return microerror.Mask(err)
		}
	}

	r.logger.Debugf(ctx, "created kubernetes client for bootstrap cluster")

	return nil
}
