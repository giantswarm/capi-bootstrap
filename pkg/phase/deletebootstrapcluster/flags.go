package deletebootstrapcluster

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagKindClusterName = "kind-cluster-name"
)

type flags struct {
	KindClusterName string
}

func (f *flags) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.KindClusterName, flagKindClusterName, "capi-bootstrap", "")
}

func (f *flags) Validate() error {
	if f.KindClusterName == "" {
		return microerror.Maskf(invalidFlagError, "--%s is required", flagKindClusterName)
	}

	return nil
}
