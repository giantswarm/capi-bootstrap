package pivot

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagClusterNamespace      = "cluster-namespace"
	flagKindClusterName       = "kind-cluster-name"
	flagManagementClusterName = "management-cluster-name"

	flagFile = "file"

	flagFromInCluster  = "from-in-cluster"
	flagFromKubeconfig = "from-kubeconfig"

	flagToInCluster  = "to-in-cluster"
	flagToKubeconfig = "to-kubeconfig"
)

type flags struct {
	ClusterNamespace      string
	KindClusterName       string
	ManagementClusterName string

	FileInputs string

	FromInCluster  bool
	FromKubeconfig string

	ToInCluster  bool
	ToKubeconfig string
}

func (f *flags) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.ClusterNamespace, flagClusterNamespace, "", "")
	cmd.Flags().StringVar(&f.KindClusterName, flagKindClusterName, "", "")
	cmd.Flags().StringVar(&f.ManagementClusterName, flagManagementClusterName, "", "")

	cmd.Flags().StringVarP(&f.FileInputs, flagFile, "f", "", `Existing release upon which to base the new release. Must follow semver format.`)

	cmd.Flags().BoolVar(&f.FromInCluster, flagFromInCluster, false, "")
	cmd.Flags().StringVar(&f.FromKubeconfig, flagFromKubeconfig, "", "")

	cmd.Flags().BoolVar(&f.ToInCluster, flagToInCluster, false, "")
	cmd.Flags().StringVar(&f.ToKubeconfig, flagToKubeconfig, "", "")
}

func (f *flags) Validate() error {
	if f.FromKubeconfig == "" == f.FromInCluster {
		return microerror.Maskf(invalidFlagError, "only one of --%s or --%s may be used", flagFromKubeconfig, flagFromInCluster)
	}
	if f.ToKubeconfig == "" == f.ToInCluster {
		return microerror.Maskf(invalidFlagError, "only one of --%s or --%s may be used", flagToKubeconfig, flagToInCluster)
	}

	return nil
}
