package init

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/cmd/init/phase/preflight"
	"github.com/giantswarm/capi-bootstrap/pkg/config"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/createbootstrapcluster"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/launch"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/uploadconfig"
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
	bootstrapConfig, err := config.FromFile(r.flag.ConfigFile)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		_, runner, err := preflight.New(preflight.Config{
			Logger: r.logger,
			Stderr: r.stderr,
			Stdout: r.stdout,
		})
		err = runner.Do(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		_, runner, err := createbootstrapcluster.New(createbootstrapcluster.Config{
			Logger: r.logger,

			Stderr: r.stderr,
			Stdout: r.stdout,
		})
		err = runner.Do(ctx, bootstrapConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		_, runner, err := uploadconfig.New(uploadconfig.Config{
			Logger: r.logger,

			Stderr: r.stderr,
			Stdout: r.stdout,
		})
		err = runner.Do(ctx, bootstrapConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		_, runner, err := launch.New(launch.Config{
			Logger: r.logger,

			Stderr: r.stderr,
			Stdout: r.stdout,
		})
		err = runner.Do(ctx, bootstrapConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
