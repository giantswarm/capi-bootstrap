package encrypt

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagClusterName = "cluster-name"
	flagInputFile   = "input-file"
)

type flags struct {
	ClusterName string
	InputFile   string
}

func (f *flags) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.ClusterName, flagClusterName, "", `Management cluster name (optional). If provided and SOPS_AGE_KEY environment variable is not defined, it will be used to look up the encryption key in Lastpass.`)
	cmd.Flags().StringVar(&f.InputFile, flagInputFile, "", `Path to file to decrypt`)
}

func (f *flags) Validate() error {
	if f.InputFile == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagInputFile)
	}

	return nil
}
