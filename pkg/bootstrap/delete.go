package bootstrap

import (
	"context"
	"io"
	"strings"

	core "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (b *Bootstrapper) Delete(ctx context.Context) error {
	defer b.cleanup()
	return b.delete(ctx)
}

func (b *Bootstrapper) delete(ctx context.Context) error {
	// 0. setup helm
	err := b.configureHelmCatalogRepo()
	if err != nil {
		return err
	}

	// 1. create/find kind cluster and get kubeconfig/client
	err = b.ensureBootstrapCluster(ctx)
	if err != nil {
		return err
	}
	/*
		// 2. setup app platform and capi on bootstrap cluster
		err = b.setupMC(ctx, false)
		if err != nil {
			return err
		}

		// 3. get kubeconfig for permanent cluster to be deleted
		err = b.loadPermanentKubeconfigFromLastPass(ctx)
		if err != nil {
			return err
		}

		// 4. move cluster resources from permanent to bootstrap cluster
		err = b.moveCluster(ctx, false)
		if err != nil {
			return err
		}
		*

	*/
	err = b.deleteCluster(ctx)
	if err != nil {
		return err
	}

	err = b.deleteBootstrapCluster()
	if err != nil {
		return err
	}

	return nil
}

func (b *Bootstrapper) deleteCluster(ctx context.Context) error {
	k8sClient := b.getClient(false)

	// 0. read cluster resources from input
	clusterResources, err := decodeObjects(io.NopCloser(strings.NewReader(b.fileInputs)))
	if err != nil {
		return err
	}

	// 1. delete cluster resources
	err = deleteResources(ctx, k8sClient, clusterResources)
	if err != nil {
		return err
	}

	// 2. wait for apps and cluster to actually be deleted
	err = waitForClusterDeleted(ctx, k8sClient, client.ObjectKey{
		Name:      b.managementClusterName,
		Namespace: b.clusterNamespace,
	})
	if err != nil {
		return err
	}

	return nil
}

func deletePVCs(ctx context.Context, ctrlClient client.Client) error {
	var claims core.PersistentVolumeClaimList
	err := ctrlClient.List(ctx, &claims)
	if err != nil {
		return err
	}
	for _, claim := range claims.Items {
		err = ctrlClient.Delete(ctx, &claim)
		if err != nil {
			return err
		}
	}
	return nil
}
