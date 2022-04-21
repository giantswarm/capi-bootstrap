package initdefaults

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

const (
	name        = "init-defaults"
	description = `Generate a bootstrap config file which can be passed into "config apply" and other bootstrap commands`
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

	var flags Flags

	r := &Runner{
		flags:  &flags,
		logger: config.Logger,
		stderr: config.Stderr,
		stdout: config.Stdout,
	}

	command := &cobra.Command{
		Use:   name,
		Short: description,
		Long:  description,
		RunE:  r.Run,
	}

	flags.Init(command)

	return command, nil
}
