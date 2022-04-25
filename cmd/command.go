package cmd

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	keycmd "github.com/giantswarm/capi-bootstrap/cmd/key"
	secretcmd "github.com/giantswarm/capi-bootstrap/cmd/secret"
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

	var keyCmd *cobra.Command
	{
		var err error
		keyCmd, err = keycmd.New(keycmd.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var secretCmd *cobra.Command
	{
		var err error
		secretCmd, err = secretcmd.New(secretcmd.Config{
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
		Use:           project.Name(),
		Short:         project.Description(),
		Long:          project.Description(),
		RunE:          runner.Run,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	command.AddCommand(keyCmd)
	command.AddCommand(secretCmd)

	return &command, nil
}
