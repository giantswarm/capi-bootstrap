package installcapicontrollers

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/capi-bootstrap/pkg/kubernetes"
	"github.com/giantswarm/capi-bootstrap/pkg/shell"
)

func (r *Runner) Run(cmd *cobra.Command, _ []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context())
	return microerror.Mask(err)
}

func (r *Runner) Do(ctx context.Context) error {
	k8sClient, err := kubernetes.ClientFromFlags(r.flag.Kubeconfig, r.flag.InCluster)
	if err != nil {
		return microerror.Mask(err)
	}
	k8sClient.Logger = r.logger

	r.logger.Debugf(ctx, "installing capi controllers")

	{
		_, stdErr, err := shell.Execute(shell.Command{
			Name: "clusterctl",
			Args: []string{
				"init",
				"--kubeconfig",
				r.flag.Kubeconfig,
				"--infrastructure",
				r.flag.Provider,
			},
		})
		if err != nil {
			return fmt.Errorf("%w: %s", err, stdErr)
		}

		var deploymentKeys []client.ObjectKey
		for _, controller := range []string{
			"capi-kubeadm-bootstrap",
			"capi-kubeadm-control-plane",
			"capi",
			"capo",
		} {
			deploymentKeys = append(deploymentKeys, client.ObjectKey{
				Namespace: fmt.Sprintf("%s-system", controller),
				Name:      fmt.Sprintf("%s-controller-manager", controller),
			})
		}
		err = k8sClient.WaitForDeployments(ctx, deploymentKeys)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "ensured CAPI controllers")
	}

	return nil
}
