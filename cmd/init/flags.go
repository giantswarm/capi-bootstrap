package init

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
	cmd.Flags().StringVar(&f.ConfigFile, flagConfigFile, "", "")
}

func (f *flags) Validate() error {
	if f.ConfigFile == "" {
		return microerror.Maskf(invalidFlagError, "--%s is required", flagConfigFile)
	}

	return nil
}
