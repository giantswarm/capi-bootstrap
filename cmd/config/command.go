package config

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	applycmd "github.com/giantswarm/capi-bootstrap/cmd/config/apply"
	deletecmd "github.com/giantswarm/capi-bootstrap/cmd/config/delete"
)

const (
	name        = "config"
	description = `Commands for managing local config files and config stored in repositories`
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

	var applyCmd *cobra.Command
	{
		var err error
		applyCmd, err = applycmd.New(applycmd.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deleteCmd *cobra.Command
	{
		var err error
		deleteCmd, err = deletecmd.New(deletecmd.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, microerror.Mask(err)
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

	command.AddCommand(applyCmd)
	command.AddCommand(deleteCmd)

	return &command, nil
}
