package preflight

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
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
	r.logger.Debugf(ctx, "running destroy preflight checks")

	// TODO: implement

	r.logger.Debugf(ctx, "completed destroy preflight checks")

	return nil
}
