package encrypt

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagInputFile = "input-file"
	flagPublicKey = "public-key"
)

type flags struct {
	InputFile string
	PublicKey   string
}

func (f *flags) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.InputFile, flagInputFile, "", `Path to file to encrypt`)
	cmd.Flags().StringVar(&f.PublicKey, flagPublicKey, "", `Public encryption key`)
}

func (f *flags) Validate() error {
	if f.InputFile == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagInputFile)
	}
	if f.PublicKey == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagPublicKey)
	}

	return nil
}
