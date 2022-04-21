package phase

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/cmd/destroy/phase/deletepermanentcluster"
	"github.com/giantswarm/capi-bootstrap/cmd/destroy/phase/preflight"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/createbootstrapcluster"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/deletebootstrapcluster"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/installappplatform"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/installcapicontrollers"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/installcertmanager"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/installclusterappsoperator"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/pivot"
	"github.com/giantswarm/capi-bootstrap/pkg/phase/uploadconfig"
)

const (
	name        = "phase"
	description = ``
)

func New(config Config) (*cobra.Command, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
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
			return nil, microerror.Mask(err)
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
			return nil, microerror.Mask(err)
		}
	}

	var deletePermanentClusterCommand *cobra.Command
	{
		var err error
		deletePermanentClusterCommand, err = deletepermanentcluster.New(deletepermanentcluster.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, microerror.Mask(err)
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
			return nil, microerror.Mask(err)
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
			return nil, microerror.Mask(err)
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
			return nil, microerror.Mask(err)
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
			return nil, microerror.Mask(err)
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
			return nil, microerror.Mask(err)
		}
	}

	var preflightCommand *cobra.Command
	{
		var err error
		preflightCommand, err = preflight.New(preflight.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, microerror.Mask(err)
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
			return nil, microerror.Mask(err)
		}
	}

	r := &Runner{
		logger: config.Logger,
		stderr: config.Stderr,
		stdout: config.Stdout,
	}

	c := &cobra.Command{
		Use:   name,
		Short: description,
		Long:  description,
		RunE:  r.Run,
	}

	c.AddCommand(createBootstrapClusterCommand)
	c.AddCommand(deleteBootstrapClusterCommand)
	c.AddCommand(deletePermanentClusterCommand)
	c.AddCommand(installAppPlatformCommand)
	c.AddCommand(installCAPIControllersCommand)
	c.AddCommand(installCertManagerCommand)
	c.AddCommand(installClusterAppsOperatorCommand)
	c.AddCommand(pivotCommand)
	c.AddCommand(preflightCommand)
	c.AddCommand(uploadConfigCommand)

	return c, nil
}
