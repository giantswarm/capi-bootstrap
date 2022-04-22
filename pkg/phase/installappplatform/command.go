package installappplatform

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	config2 "github.com/giantswarm/capi-bootstrap/pkg/config"
)

const (
	name        = "install-app-platform"
	description = `Install Giant Swarm app platform`
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

	var flags config2.Flag

	runner := Runner{
		flag:   &flags,
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

	flags.Init(&command)

	return &command, &runner, nil
}
