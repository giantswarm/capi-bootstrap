package apply

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagManagementClusterName = "name"
)

type flags struct {
	ManagementClusterName string
}

func (f *flags) Init(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.ManagementClusterName, flagManagementClusterName, "n", "", `Management cluster name`)
}

func (f *flags) Validate() error {
	if f.ManagementClusterName == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagManagementClusterName)
	}

	return nil
}
