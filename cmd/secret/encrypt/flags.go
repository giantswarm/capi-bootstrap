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
	cmd.Flags().StringVar(&f.InputFile, flagClusterName, "", `Management cluster name (optional). If SOPS_AGE_RECEIPIENTS environment variable is not provided, the cluster name can be used to look up the encryption key from Lastpass.`)
	cmd.Flags().StringVar(&f.InputFile, flagInputFile, "", `Path to file to encrypt`)
}

func (f *flags) Validate() error {
	if f.InputFile == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagInputFile)
	}

	return nil
}
