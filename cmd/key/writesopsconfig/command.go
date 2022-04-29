package writesopsconfig

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

const (
	name        = "write-sops-config"
	description = `Generate a sops config file (see https://github.com/mozilla/sops#using-sops-yaml-conf-to-select-kms-pgp-for-new-files) based on the provided SOPS_AGE_RECIPIENTS environment variable`
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

	return &command, nil
}
