package pivot

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/pkg/kubernetes"
	"github.com/giantswarm/capi-bootstrap/pkg/shell"
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
	r.logger.Debugf(ctx, "pivoting management cluster")

	var source ClusterScope
	{
		source.KubeconfigPath = r.flag.FromKubeconfig
		var err error
		source.K8sClient, err = kubernetes.ClientFromFlags(r.flag.FromKubeconfig, r.flag.FromInCluster)
		if err != nil {
			return microerror.Mask(err)
		}
		source.K8sClient.Logger = r.logger
	}

	var target ClusterScope
	{
		target.KubeconfigPath = r.flag.FromKubeconfig
		var err error
		target.K8sClient, err = kubernetes.ClientFromFlags(r.flag.ToKubeconfig, r.flag.ToInCluster)
		if err != nil {
			return microerror.Mask(err)
		}
		target.K8sClient.Logger = r.logger
	}

	_, stdErr, err := shell.Execute(shell.Command{
		Name: "clusterctl",
		Args: []string{
			"move",
			"--namespace",
			r.flag.ClusterNamespace,
			"--kubeconfig",
			source.KubeconfigPath,
			"--to-kubeconfig",
			target.KubeconfigPath,
		},
	})
	if err != nil {
		return microerror.Mask(fmt.Errorf("%w: %s", err, stdErr))
	}

	clusterResources, err := kubernetes.DecodeObjects(io.NopCloser(strings.NewReader(r.flag.FileInputs)))
	if err != nil {
		return microerror.Mask(err)
	}

	err = target.K8sClient.ApplyResources(ctx, clusterResources)
	if err != nil {
		return microerror.Mask(err)
	}

	err = target.K8sClient.WaitForClusterReady(ctx, r.flag.ClusterNamespace, r.flag.ManagementClusterName)
	if err != nil {
		return microerror.Mask(err)
	}

	err = source.K8sClient.DeleteResources(ctx, clusterResources)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "pivoted management cluster")

	return nil
}
