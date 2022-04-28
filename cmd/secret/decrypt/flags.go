package encrypt

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagInputFile  = "input-file"
	flagPrivateKey = "private-key"
)

type flags struct {
	InputFile string
}

func (f *flags) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.InputFile, flagInputFile, "", `Path to file to encrypt`)
}

func (f *flags) Validate() error {
	if f.InputFile == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagInputFile)
	}

	return nil
}
