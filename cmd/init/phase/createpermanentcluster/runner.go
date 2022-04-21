package createpermanentcluster

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/ansd/lastpass-go"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/pkg/kubernetes"
)

func (r *Runner) Run(cmd *cobra.Command, _ []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context())
	return microerror.Mask(err)
}

func getAccount(ctx context.Context, client *lastpass.Client, group, name string) (*lastpass.Account, error) {
	accounts, err := client.Accounts(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	for _, account := range accounts {
		if account.Name == name && account.Group == group {
			return account, nil
		}
	}
	return nil, microerror.Mask(errors.New("not found"))
}

func (r *Runner) Do(ctx context.Context) error {
	var k8sClient *kubernetes.Client
	{
		k8sClient, err := kubernetes.ClientFromFlags(r.flag.Kubeconfig, r.flag.InCluster)
		if err != nil {
			return microerror.Mask(err)
		}
		k8sClient.Logger = r.logger
	}

	lastPassClient, err := lastpass.NewClient(ctx, "username", "password", lastpass.WithOneTimePassword("123456"))
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "creating permanent cluster")

	err = k8sClient.CreateNamespace(ctx, r.flag.ClusterNamespace)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterResources, err := kubernetes.DecodeObjects(io.NopCloser(strings.NewReader(r.flag.FileInputs)))
	if err != nil {
		return microerror.Mask(err)
	}

	{
		openrcSecret, err := getAccount(ctx, lastPassClient, "Shared-Team Rocket", "openrc")
		if err != nil {
			return microerror.Mask(err)
		}

		err = k8sClient.CreateCloudConfigSecret(ctx, openrcSecret.Notes, r.flag.ClusterNamespace)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = k8sClient.ApplyResources(ctx, clusterResources)
	if err != nil {
		return microerror.Mask(err)
	}

	err = k8sClient.WaitForClusterReady(ctx, r.flag.ClusterNamespace, r.flag.ManagementClusterName)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "created permanent cluster")

	return nil
}
