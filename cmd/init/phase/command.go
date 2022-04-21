package phase

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/cmd/init/phase/createpermanentcluster"
	"github.com/giantswarm/capi-bootstrap/cmd/init/phase/preflight"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/createbootstrapcluster"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/deletebootstrapcluster"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/installappplatform"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/installcapicontrollers"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/installcertmanager"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/installclusterappsoperator"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/launch"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/pivot"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/run"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/uploadconfig"
)

const (
	name        = "phase"
	description = ``
)

func New(config Config) (*cobra.Command, *Runner, error) {
	if config.Logger == nil {
		return nil, nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Stderr == nil {
		config.Stderr = os.Stderr
	}
	if config.Stdout == nil {
		config.Stdout = os.Stdout
	}

	var createBootstrapClusterCommand *cobra.Command
	{
		var err error
		createBootstrapClusterCommand, _, err = createbootstrapcluster.New(createbootstrapcluster.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	}

	var createPermanentClusterCommand *cobra.Command
	{
		var err error
		createPermanentClusterCommand, _, err = createpermanentcluster.New(createpermanentcluster.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	}

	var deleteBootstrapClusterCommand *cobra.Command
	{
		var err error
		deleteBootstrapClusterCommand, err = deletebootstrapcluster.New(deletebootstrapcluster.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	}

	var installAppPlatformCommand *cobra.Command
	{
		var err error
		installAppPlatformCommand, _, err = installappplatform.New(installappplatform.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	}

	var installCAPIControllersCommand *cobra.Command
	{
		var err error
		installCAPIControllersCommand, _, err = installcapicontrollers.New(installcapicontrollers.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	}

	var installCertManagerCommand *cobra.Command
	{
		var err error
		installCertManagerCommand, _, err = installcertmanager.New(installcertmanager.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	}

	var installClusterAppsOperatorCommand *cobra.Command
	{
		var err error
		installClusterAppsOperatorCommand, _, err = installclusterappsoperator.New(installclusterappsoperator.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	}

	var launchCommand *cobra.Command
	{
		var err error
		launchCommand, _, err = launch.New(launch.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	}

	var pivotCommand *cobra.Command
	{
		var err error
		pivotCommand, err = pivot.New(pivot.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	}

	var preflightCommand *cobra.Command
	{
		var err error
		preflightCommand, _, err = preflight.New(preflight.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	}

	var runCommand *cobra.Command
	{
		var err error
		runCommand, _, err = run.New(run.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	}

	var uploadConfigCommand *cobra.Command
	{
		var err error
		uploadConfigCommand, _, err = uploadconfig.New(uploadconfig.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	}

	runner := Runner{
		logger: config.Logger,
		stderr: config.Stderr,
		stdout: config.Stdout,
	}

	command := cobra.Command{
		Use:   name,
		Short: description,
		Long:  description,
		RunE:  runner.Run,
	}

	command.AddCommand(createBootstrapClusterCommand)
	command.AddCommand(createPermanentClusterCommand)
	command.AddCommand(deleteBootstrapClusterCommand)
	command.AddCommand(installAppPlatformCommand)
	command.AddCommand(installCAPIControllersCommand)
	command.AddCommand(installCertManagerCommand)
	command.AddCommand(installClusterAppsOperatorCommand)
	command.AddCommand(launchCommand)
	command.AddCommand(pivotCommand)
	command.AddCommand(preflightCommand)
	command.AddCommand(runCommand)
	command.AddCommand(uploadConfigCommand)

	return &command, &runner, nil
}
