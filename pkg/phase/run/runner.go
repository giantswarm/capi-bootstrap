package run

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/cmd/init/phase/createpermanentcluster"
	"github.com/giantswarm/capi-bootstrap/cmd/init/phase/preflight"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/installappplatform"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/installcapicontrollers"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/installcertmanager"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/installclusterappsoperator"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/launch"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/uploadconfig"
)

func (r *Runner) Run(cmd *cobra.Command, _ []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context())
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Runner) Do(ctx context.Context) error {
	environment, err := r.flag.BuildEnvironment(r.logger)
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
		_, runner, err := installappplatform.New(installappplatform.Config{
			Logger: r.logger,

			Stderr: r.stderr,
			Stdout: r.stdout,
		})
		err = runner.Do(ctx, environment)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		_, runner, err := installcertmanager.New(installcertmanager.Config{
			Logger: r.logger,

			Stderr: r.stderr,
			Stdout: r.stdout,
		})
		err = runner.Do(ctx, environment)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		_, runner, err := installcapicontrollers.New(installcapicontrollers.Config{
			Logger: r.logger,

			Stderr: r.stderr,
			Stdout: r.stdout,
		})
		err = runner.Do(ctx, environment)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		_, runner, err := installclusterappsoperator.New(installclusterappsoperator.Config{
			Logger: r.logger,

			Stderr: r.stderr,
			Stdout: r.stdout,
		})
		err = runner.Do(ctx, environment)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		_, runner, err := createpermanentcluster.New(createpermanentcluster.Config{
			Logger: r.logger,

			Stderr: r.stderr,
			Stdout: r.stdout,
		})
		err = runner.Do(ctx, environment)
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
		err = runner.Do(ctx, environment)
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
		err = runner.Do(ctx, environment)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
