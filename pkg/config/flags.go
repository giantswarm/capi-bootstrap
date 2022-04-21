package config

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagConfigFile = "config-file"
)

type Flag struct {
	ConfigFile string
}

func (f *Flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.ConfigFile, flagConfigFile, "", "")
}

func (f *Flag) Validate() error {
	if f.ConfigFile == "" {
		return microerror.Maskf(invalidFlagError, "--%s is required", flagConfigFile)
	}

	return nil
}

func (f *Flag) ToConfig() (BootstrapConfig, error) {
	return FromFile(f.ConfigFile)
}
