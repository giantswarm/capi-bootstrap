package cmd

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

func (r *Runner) Run(cmd *cobra.Command, args []string) error {
	err := r.Do(cmd.Context(), cmd, args)
	return microerror.Mask(err)
}

func (r *Runner) Do(_ context.Context, cmd *cobra.Command, _ []string) error {
	err := cmd.Help()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
