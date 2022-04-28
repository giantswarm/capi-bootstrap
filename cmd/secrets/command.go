package secrets

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	generatecmd "github.com/giantswarm/capi-bootstrap/cmd/secrets/generate"
)

const (
	name        = "secrets"
	description = `Commands for managing installation-level secrets`
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

	var generateCmd *cobra.Command
	{
		var err error
		generateCmd, err = generatecmd.New(generatecmd.Config{
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

	command.AddCommand(generateCmd)

	return &command, nil
}
