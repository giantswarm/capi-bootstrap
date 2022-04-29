package key

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	decryptcmd "github.com/giantswarm/capi-bootstrap/cmd/secret/decrypt"
	encryptcmd "github.com/giantswarm/capi-bootstrap/cmd/secret/encrypt"
)

const (
	name        = "secret"
	description = `Commands for managing the encryption of Kubernetes secrets`
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

	var decryptCmd *cobra.Command
	{
		var err error
		decryptCmd, err = decryptcmd.New(decryptcmd.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var encryptCmd *cobra.Command
	{
		var err error
		encryptCmd, err = encryptcmd.New(encryptcmd.Config{
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

	command.AddCommand(decryptCmd)
	command.AddCommand(encryptCmd)

	return &command, nil
}
