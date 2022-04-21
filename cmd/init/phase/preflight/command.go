package preflight

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	config2 "github.com/giantswarm/capi-bootstrap/pkg/config"
)

const (
	name        = "preflight"
	description = `Create kind cluster to act as temporary MC during bootstrap`
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

	var flag config2.Flag

	runner := Runner{
		flag: &flag,

		Logger: config.Logger,

		Stderr: config.Stderr,
		Stdout: config.Stdout,
	}

	command := cobra.Command{
		Use:   name,
		Short: description,
		Long:  description,
		RunE:  runner.Run,
	}

	flag.Init(&command)

	return &command, &runner, nil
}
