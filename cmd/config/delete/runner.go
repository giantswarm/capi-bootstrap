package apply

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/pkg/fleet"
)

func (r *Runner) Run(cmd *cobra.Command, args []string) error {
	err := r.Do(cmd.Context(), cmd, args)
	return microerror.Mask(err)
}

func (r *Runner) Do(ctx context.Context, _ *cobra.Command, _ []string) error {
	fleetService, err := fleet.New(fleet.Config{
		ClusterManifestFile: "cluster.yaml",
		ClusterName:         "guppy",
	})
	if err != nil {
		return microerror.Mask(err)
	}

	err = fleetService.EnsureDeleted(ctx)
	return microerror.Mask(err)
}
