package apply

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagConfigFile = "config-file"
)

type flags struct {
	ConfigFile string
}

func (f *flags) Init(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.ConfigFile, flagConfigFile, "c", "", ``)
}

func (f *flags) Validate() error {
	if f.ConfigFile == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagConfigFile)
	}

	return nil
}
