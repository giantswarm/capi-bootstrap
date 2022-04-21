package preflight

import (
	"context"
	"log"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/pkg/shell"
)

func (r *Runner) Run(cmd *cobra.Command, _ []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context())
	return microerror.Mask(err)
}

func (r *Runner) Do(ctx context.Context) error {
	r.Logger.Debugf(ctx, "running init preflight checks")

	if os.Getenv("GITHUB_TOKEN") == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	err := shell.VerifyBinaryExists("helm")
	if err != nil {
		return microerror.Mask(err)
	}

	err = shell.VerifyBinaryExists("lpass")
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
