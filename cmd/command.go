package cmd

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	configcmd "github.com/giantswarm/capi-bootstrap/cmd/config"
	destroycmd "github.com/giantswarm/capi-bootstrap/cmd/destroy"
	initcmd "github.com/giantswarm/capi-bootstrap/cmd/init"
	"github.com/giantswarm/capi-bootstrap/pkg/project"
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

	var configCmd *cobra.Command
	{
		var err error
		configCmd, err = configcmd.New(configcmd.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var destroyCmd *cobra.Command
	{
		var err error
		destroyCmd, err = destroycmd.New(destroycmd.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var initCmd *cobra.Command
	{
		var err error
		initCmd, _, err = initcmd.New(initcmd.Config{
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
		Use:           project.Name(),
		Short:         project.Description(),
		Long:          project.Description(),
		RunE:          r.Run,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	c.AddCommand(configCmd)
	c.AddCommand(destroyCmd)
	c.AddCommand(initCmd)

	return c, nil
}
