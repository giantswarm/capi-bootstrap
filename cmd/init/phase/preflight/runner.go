package preflight

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/pkg/shell"
)

func (r *Runner) Run(cmd *cobra.Command, _ []string) error {
	err := r.Do(cmd.Context())
	return microerror.Mask(err)
}

func (r *Runner) Do(ctx context.Context) error {
	r.Logger.Debugf(ctx, "running init preflight checks")

	err := shell.VerifyBinaryExists("helm")
	if err != nil {
		return microerror.Mask(err)
	}

	err = shell.VerifyBinaryExists("kind")
	if err != nil {
		return microerror.Mask(err)
	}

	err = shell.VerifyBinaryExists("opsctl")
	if err != nil {
		return microerror.Mask(err)
	}

	r.Logger.Debugf(ctx, "init preflight checks passed")

	return nil
}
