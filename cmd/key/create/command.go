package create

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

const (
	name        = "create"
	description = `Create a management cluster encryption key in the password manager fetching an existing key if it already exists. Will print the public and private key as environment variables SOPS_AGE_RECIPIENTS and SOPS_AGE_KEY respectively for use by other commands.`
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

	runner := Runner{
		flag:   &flags{},
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

	runner.flag.Init(&command)

	return &command, nil
}
