package init

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/cmd/init/phase"
)

const (
	name        = "init"
	description = `Top-level command for creating a new CAPI MC`
)

func New(config Config) (*cobra.Command, *Runner, error) {
	if config.Logger == nil {
		return nil, nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Stderr == nil {
		config.Stderr = os.Stderr
	}
	if config.Stdout == nil {
		config.Stdout = os.Stdout
	}

	var phaseCommand *cobra.Command
	{
		var err error
		phaseCommand, _, err = phase.New(phase.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		})
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	}

	var flag flags

	runner := Runner{
		flag: &flag,

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

	flag.Init(&command)

	command.AddCommand(phaseCommand)

	return &command, &runner, nil
}
