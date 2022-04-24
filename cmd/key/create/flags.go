package create

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagClusterName = "cluster-name"
)

type flags struct {
	ClusterName string
}

func (f *flags) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.ClusterName, flagClusterName, "", `Management cluster name`)
}

func (f *flags) Validate() error {
	if f.ClusterName == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagClusterName)
	}

	return nil
}
