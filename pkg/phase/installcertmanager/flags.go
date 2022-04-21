package installcertmanager

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagInCluster  = "in-cluster"
	flagKubeconfig = "kubeconfig"
)

type flags struct {
	InCluster  bool
	Kubeconfig string
}

func (f *flags) Init(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&f.InCluster, flagInCluster, false, "")
	cmd.Flags().StringVar(&f.Kubeconfig, flagKubeconfig, "", "")
}

func (f *flags) Validate() error {
	if f.Kubeconfig == "" != f.InCluster {
		return microerror.Maskf(invalidFlagError, "only one of --%s or --%s may be used", flagKubeconfig, flagInCluster)
	}

	return nil
}
