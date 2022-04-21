package deletepermanentcluster

import (
	"context"
	"io"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/pkg/kubernetes"
)

func (r *Runner) Run(cmd *cobra.Command, args []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context(), cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Runner) Do(ctx context.Context, _ *cobra.Command, _ []string) error {
	k8sClient, err := kubernetes.ClientFromFlags(r.flag.Kubeconfig, r.flag.InCluster)
	if err != nil {
		return microerror.Mask(err)
	}
	k8sClient.Logger = r.logger

	r.logger.Debugf(ctx, "deleting permanent cluster")

	clusterResources, err := kubernetes.DecodeObjects(io.NopCloser(strings.NewReader(r.flag.FileInputs)))
	if err != nil {
		return microerror.Mask(err)
	}

	err = k8sClient.DeleteResources(ctx, clusterResources)
	if err != nil {
		return microerror.Mask(err)
	}

	err = k8sClient.WaitForClusterDeleted(ctx, r.flag.ClusterNamespace, r.flag.ManagementClusterName)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "deleted permanent cluster")

	return nil
}
