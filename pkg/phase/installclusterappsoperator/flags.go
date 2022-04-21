package installclusterappsoperator

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagInCluster  = "in-cluster"
	flagKubeconfig = "kubeconfig"

	flagBaseDomain = "base-domain"
	flagProvider   = "provider"
)

type flags struct {
	InCluster  bool
	Kubeconfig string

	BaseDomain string
	Provider   string
}

func (f *flags) Init(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&f.InCluster, flagInCluster, false, "")
	cmd.Flags().StringVar(&f.Kubeconfig, flagKubeconfig, "", "")

	cmd.Flags().StringVar(&f.BaseDomain, flagBaseDomain, "", "")
	cmd.Flags().StringVar(&f.Provider, flagProvider, "", "")
}

func (f *flags) Validate() error {
	if f.Kubeconfig == "" != f.InCluster {
		return microerror.Maskf(invalidFlagError, "only one of --%s or --%s may be used", flagKubeconfig, flagInCluster)
	}

	return nil
}
