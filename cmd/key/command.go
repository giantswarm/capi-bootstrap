package key

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	createcmd "github.com/giantswarm/capi-bootstrap/cmd/key/create"
	deletecmd "github.com/giantswarm/capi-bootstrap/cmd/key/delete"
)

const (
	name        = "key"
	description = `Commands for managing encryption keys`
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

	var createCmd *cobra.Command
	{
		var err error
		createCmd, err = createcmd.New(createcmd.Config{
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

	command.AddCommand(createCmd)
	command.AddCommand(deleteCmd)

	return &command, nil
}
