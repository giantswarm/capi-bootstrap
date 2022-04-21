package preflight

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagInCluster  = "in-cluster"
	flagKubeconfig = "kubeconfig"
)

type Flag struct {
	InCluster  bool
	Kubeconfig string

	CloudConfigSecret     string
	ClusterNamespace      string
	FileInputs            string
	ManagementClusterName string
}

func (f *Flag) Init(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&f.InCluster, flagInCluster, false, "")
	cmd.Flags().StringVar(&f.Kubeconfig, flagKubeconfig, "", "")
}

func (f *Flag) Validate() error {
	if f.Kubeconfig == "" {
		return microerror.Maskf(invalidFlagError, "--%s is required", flagKubeconfig)
	}

	return nil
}
