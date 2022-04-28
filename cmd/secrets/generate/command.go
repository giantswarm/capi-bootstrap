package generate

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

const (
	name        = "generate"
	description = `Generates configuration files for a new management cluster`
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
