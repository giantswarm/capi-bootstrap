package bootstrap

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	application "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// 2. setup app platform and capi on bootstrap cluster
	err = b.setupMC(ctx, false)
	if err != nil {
		return err
	}

	// 3. get kubeconfig for permanent cluster to be deleted
	err = b.loadPermanentKubeconfigFromLastPass()
	if err != nil {
		return err
	}

	// 4. move cluster resources from permanent to bootstrap cluster
	err = b.moveCluster(ctx, false)
	if err != nil {
		return err
	}

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

// TODO: use this function
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

func waitForClusterDeleted(ctx context.Context, k8sClient client.Client, clusterKey client.ObjectKey) error {
	for {
		var clusterApps application.AppList
		err := k8sClient.List(ctx, &clusterApps, client.InNamespace(clusterKey.Namespace), client.MatchingLabels{
			"giantswarm.io/cluster": clusterKey.Name,
		})
		if err != nil {
			return err
		}

		if len(clusterApps.Items) == 0 {
			fmt.Println("all apps deleted")
			break
		}

		appsByName := map[string]application.App{}
		for _, app := range clusterApps.Items {
			appName := strings.TrimPrefix(app.ObjectMeta.Name, fmt.Sprintf("%s-", clusterKey.Name))
			if appName != "" {
				appsByName[appName] = app
			}
		}

		allAppsDeleted := true
		for _, appName := range []string{
			"app-operator",
			"chart-operator",
			"cert-exporter",
			"cilium",
			"cloud-provider-openstack",
			"kube-state-metrics",
			"metrics-server",
			"net-exporter",
			"node-exporter",
		} {
			if _, ok := appsByName[appName]; ok {
				fmt.Printf("waiting for app %s to be deleted\n", appName)
				allAppsDeleted = false
				break
			}
		}

		if allAppsDeleted {
			break
		}

		time.Sleep(time.Second)
	}

	fmt.Print("waiting for cluster to be deleted")

	for {
		cluster := meta.PartialObjectMetadata{
			TypeMeta: meta.TypeMeta{
				Kind:       "Cluster",
				APIVersion: "cluster.x-k8s.io/v1beta1",
			},
		}
		err := k8sClient.Get(ctx, clusterKey, &cluster)
		if apierrors.IsNotFound(err) {
			// deleted
			fmt.Print("\n")
			break
		} else if err != nil {
			return err
		}

		fmt.Print(".")
		time.Sleep(time.Second)
	}

	return nil
}

func (b *Bootstrapper) deleteBootstrapCluster() error {
	if exists, err := b.kindClient.ClusterExists(b.kindClusterName); err != nil {
		return err
	} else if exists {
		err = b.kindClient.DeleteCluster(b.kindClusterName)
		if err != nil {
			return err
		}
	}

	return nil
}
