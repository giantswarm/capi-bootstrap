package bootstrap

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	application "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/shell"
	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

func (b *Bootstrapper) Create(ctx context.Context) error {
	defer b.cleanup()
	return b.create(ctx)
}

func (b *Bootstrapper) create(ctx context.Context) error {
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

	// 3. create permanent cluster as apps in bootstrap cluster and wait for ready
	err = b.createCluster(ctx)
	if err != nil {
		return err
	}

	// 4. setup app platform and capi on permanent cluster
	err = b.setupMC(ctx, true)
	if err != nil {
		return err
	}

	// 5. move cluster resources from bootstrap into permanent cluster
	err = b.moveCluster(ctx, true)
	if err != nil {
		return err
	}

	return nil
}

func waitForClusterReady(ctx context.Context, k8sClient client.Client, clusterKey client.ObjectKey) error {
	for {
		var clusterApps application.AppList
		err := k8sClient.List(ctx, &clusterApps, client.InNamespace(clusterKey.Namespace), client.MatchingLabels{
			"giantswarm.io/cluster": clusterKey.Name,
		})
		if err != nil {
			return err
		}

		if len(clusterApps.Items) == 0 {
			fmt.Println("waiting for cluster apps to be created")
			time.Sleep(time.Second)
			continue
		}

		appsByName := map[string]application.App{}
		for _, app := range clusterApps.Items {
			appName := strings.TrimPrefix(app.ObjectMeta.Name, fmt.Sprintf("%s-", clusterKey.Name))
			if appName != "" {
				appsByName[appName] = app
			}
		}

		allAppsDeployed := true
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
			if app, ok := appsByName[appName]; !ok {
				fmt.Printf("waiting for app %s to be created\n", appName)
				allAppsDeployed = false
				break
			} else if app.Status.Release.Status != "deployed" {
				fmt.Printf("waiting for app %s to have status \"deployed\", current status \"%s\"\n", appName, app.Status.Release.Status)
				allAppsDeployed = false
				break
			}
		}

		if allAppsDeployed {
			break
		}

		time.Sleep(time.Second)
	}

	return nil
}

func (b *Bootstrapper) createCloudConfigSecret(ctx context.Context, k8sClient client.Client, clusterResources []client.Object) error {
	cloudConfigName, err := extractClusterCloudConfigName(clusterResources)
	if err != nil {
		return err
	}

	project := strings.TrimPrefix(cloudConfigName, "cloud-config-")
	openrcSecret, err := b.lastPassClient.GetSecret("Shared-Customers/THG", fmt.Sprintf("%s-openrc.sh", project))
	if err != nil {
		return err
	}

	cloudConfig := openrcToCloudConfig(openrcSecret.Note)
	cloudConfigYAML, err := yaml.Marshal(cloudConfig)
	if err != nil {
		return err
	}

	secret := core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Labels: map[string]string{
				"clusterctl.cluster.x-k8s.io/move": "true",
			},
			Name:      cloudConfigName,
			Namespace: b.clusterNamespace,
		},
		StringData: map[string]string{
			"clouds.yaml": string(cloudConfigYAML),
		},
	}

	err = applyResources(ctx, k8sClient, []client.Object{&secret})
	if err != nil {
		return err
	}

	return nil
}

func (b *Bootstrapper) createCluster(ctx context.Context) error {
	k8sClient := b.getClient(false)

	// 0. ensure cluster namespace exists
	err := createNamespace(ctx, k8sClient, b.clusterNamespace)
	if err != nil {
		return err
	}

	// 1. read cluster template as kubernetes objects
	clusterResources, err := decodeObjects(io.NopCloser(strings.NewReader(b.fileInputs)))
	if err != nil {
		return err
	}

	// 2. create cloud config secret before creating cluster
	err = b.createCloudConfigSecret(ctx, k8sClient, clusterResources)
	if err != nil {
		return err
	}

	// 3. create cluster resources
	err = applyResources(ctx, k8sClient, clusterResources)
	if err != nil {
		return err
	}

	// 4. wait for cluster to become ready
	err = waitForClusterReady(ctx, k8sClient, client.ObjectKey{
		Name:      b.managementClusterName,
		Namespace: b.clusterNamespace,
	})
	if err != nil {
		return err
	}

	// 5. read kubeconfig for created "permanent" cluster
	err = b.loadPermanentKubeconfigFromBootstrap(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bootstrapper) setupMC(ctx context.Context, permanent bool) error {
	err := b.installAppPlatform(ctx, permanent)
	if err != nil {
		return err
	}

	err = b.installCAPIControllers(ctx, permanent)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bootstrapper) ensureBootstrapCluster(ctx context.Context) error {
	kubeconfigFile, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		return err
	}

	b.bootstrap.KubeconfigPath = kubeconfigFile.Name()

	var kubeconfigData []byte
	if exists, err := b.kindClient.ClusterExists(b.kindClusterName); err != nil {
		return err
	} else if exists {
		kubeconfigData, err = b.kindClient.GetKubeconfig(b.kindClusterName)
		if err != nil {
			return err
		}

		err = os.WriteFile(b.bootstrap.KubeconfigPath, kubeconfigData, 0644)
		if err != nil {
			return err
		}
	} else {
		kubeconfigData, err = b.kindClient.CreateCluster(b.kindClusterName, b.bootstrap.KubeconfigPath)
		if err != nil {
			return err
		}
	}

	b.bootstrap.K8sClient, err = util.KubeconfigToClient(kubeconfigData)
	if err != nil {
		return err
	}

	return nil
}


func (b *Bootstrapper) installAppPlatform(ctx context.Context, permanent bool) error {
	k8sClient := b.getClient(permanent)
	kubeconfigPath := b.getKubeconfigPath(permanent)

	// 1. create giantswarm namespace
	err := createNamespace(ctx, k8sClient, "giantswarm")
	if err != nil {
		return err
	}

	// 2. create CRDs in application.giantswarm.io group
	{
		crds, err := b.fetchAppPlatformCRDs(ctx)
		if err != nil {
			return err
		}

		err = applyResources(ctx, k8sClient, crds)
		if err != nil {
			return err
		}
	}

	// 4. install chart-operator with helm
	_, stdErr, err := shell.Execute(shell.Command{
		Name: "helm",
		Args: []string{
			"upgrade",
			"--install",
			"--namespace",
			"giantswarm",
			"--kubeconfig",
			kubeconfigPath,
			"chart-operator",
			"control-plane-catalog/chart-operator",
			"--set",
			"chartOperator.cni.install=true",
		},
	})
	if err != nil {
		return fmt.Errorf("%q: %s", err, stdErr)
	}

	// 5. install app-operator with helm
	_, stdErr, err = shell.Execute(shell.Command{
		Name: "helm",
		Args: []string{
			"upgrade",
			"--install",
			"--namespace",
			"giantswarm",
			"--kubeconfig",
			kubeconfigPath,
			"app-operator",
			"control-plane-catalog/app-operator",
			"--set",
			fmt.Sprintf("provider.kind=%s", b.provider),
		},
	})
	if err != nil {
		return fmt.Errorf("%q: %s", err, stdErr)
	}

	// 6. install catalogs
	{
		branchName := b.managementClusterName + "_auto_branch"

		err = b.createInstallationsBranch(ctx, branchName)
		if err != nil {
			return err
		}

		err = createNamespace(ctx, k8sClient, "draughtsman")
		if err != nil {
			return err
		}

		resources := []client.Object{
			&core.ConfigMap{
				ObjectMeta: meta.ObjectMeta{
					Name:      "draughtsman-values-configmap",
					Namespace: "draughtsman",
				},
			},
			&core.Secret{
				ObjectMeta: meta.ObjectMeta{
					Name:      "draughtsman-values-secret",
					Namespace: "draughtsman",
				},
			},
		}

		err = applyResources(ctx, k8sClient, resources)
		if err != nil {
			return err
		}

		err = b.opsctlClient.EnsureCatalogs(b.managementClusterName, branchName, kubeconfigPath)
		if err != nil {
			return err
		}
	}

	// 7. install apps to adopt helm releases
	{
		apps := []client.Object{
			&core.ConfigMap{
				ObjectMeta: meta.ObjectMeta{
					Name:      "app-operator-user-values",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values": fmt.Sprintf(`provider:
  kind: %s
`, b.provider),
				},
			},
			&core.ConfigMap{
				ObjectMeta: meta.ObjectMeta{
					Name:      "chart-operator-user-values",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values": `chartOperator:
  cni:
    install: true
`,
				},
			},
			&core.ConfigMap{
				ObjectMeta: meta.ObjectMeta{
					Name:      "cluster-apps-operator-user-values",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values": fmt.Sprintf(`baseDomain: %s
provider:
  kind: %s
`, b.baseDomain, b.provider),
				},
			},
			&application.App{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"chart-operator.giantswarm.io/force-helm-upgrade": "true",
					},
					Labels: map[string]string{
						"app-operator.giantswarm.io/version": "0.0.0",
					},
					Name:      "app-operator",
					Namespace: "giantswarm",
				},
				Spec: application.AppSpec{
					Catalog:          "control-plane-catalog",
					CatalogNamespace: "giantswarm",
					KubeConfig: application.AppSpecKubeConfig{
						InCluster: true,
					},
					Name:      "app-operator",
					Namespace: "giantswarm",
					UserConfig: application.AppSpecUserConfig{
						ConfigMap: application.AppSpecUserConfigConfigMap{
							Name:      "app-operator-user-values",
							Namespace: "giantswarm",
						},
					},
					Version: "5.8.0",
				},
			},
			&application.App{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"chart-operator.giantswarm.io/force-helm-upgrade": "true",
					},
					Labels: map[string]string{
						"app-operator.giantswarm.io/version": "0.0.0",
					},
					Name:      "chart-operator",
					Namespace: "giantswarm",
				},
				Spec: application.AppSpec{
					Catalog:          "control-plane-catalog",
					CatalogNamespace: "giantswarm",
					KubeConfig: application.AppSpecKubeConfig{
						InCluster: true,
					},
					Name:      "chart-operator",
					Namespace: "giantswarm",
					UserConfig: application.AppSpecUserConfig{
						ConfigMap: application.AppSpecUserConfigConfigMap{
							Name:      "chart-operator-user-values",
							Namespace: "giantswarm",
						},
					},
					Version: "2.20.1",
				},
			},
			&application.App{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"chart-operator.giantswarm.io/force-helm-upgrade": "true",
					},
					Labels: map[string]string{
						"app-operator.giantswarm.io/version": "0.0.0",
					},
					Name:      "cluster-apps-operator",
					Namespace: "giantswarm",
				},
				Spec: application.AppSpec{
					Catalog:          "control-plane-catalog",
					CatalogNamespace: "giantswarm",
					KubeConfig: application.AppSpecKubeConfig{
						InCluster: true,
					},
					Name:      "cluster-apps-operator",
					Namespace: "giantswarm",
					UserConfig: application.AppSpecUserConfig{
						ConfigMap: application.AppSpecUserConfigConfigMap{
							Name:      "cluster-apps-operator-user-values",
							Namespace: "giantswarm",
						},
					},
					Version: "1.5.0",
				},
			},
		}

		err = applyResources(ctx, k8sClient, apps)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Bootstrapper) installCAPIControllers(ctx context.Context, permanent bool) error {
	k8sClient := b.getClient(permanent)
	kubeconfigPath := b.getKubeconfigPath(permanent)

	// 0. install cert-manager
	{
		apps := []client.Object{
			&application.App{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"chart-operator.giantswarm.io/force-helm-upgrade": "true",
					},
					Labels: map[string]string{
						"app-operator.giantswarm.io/version": "0.0.0",
					},
					Name:      "cert-manager",
					Namespace: "giantswarm",
				},
				Spec: application.AppSpec{
					Catalog:          "control-plane-catalog",
					CatalogNamespace: "giantswarm",
					KubeConfig: application.AppSpecKubeConfig{
						InCluster: true,
					},
					Name:      "cert-manager-app",
					Namespace: "giantswarm",
					Version:   "2.12.0",
				},
			},
		}

		err := applyResources(ctx, k8sClient, apps)
		if err != nil {
			return err
		}

		var deploymentKeys []client.ObjectKey
		for _, name := range []string{
			"cert-manager-controller",
			"cert-manager-webhook",
			"cert-manager-cainjector",
		} {
			deploymentKeys = append(deploymentKeys, client.ObjectKey{
				Namespace: "giantswarm",
				Name:      name,
			})
		}
		err = waitForDeployments(ctx, k8sClient, deploymentKeys)
		if err != nil {
			return err
		}

		err = waitForCRDs(ctx, k8sClient, []string{
			"certificaterequests.cert-manager.io",
			"certificates.cert-manager.io",
			"clusterissuers.cert-manager.io",
			"issuers.cert-manager.io",
		})
		if err != nil {
			return err
		}
	}

	// 1. clusterctl init
	{
		var capiNamespace core.Namespace
		err := k8sClient.Get(ctx, client.ObjectKey{Name: "capi-system"}, &capiNamespace)
		if apierrors.IsNotFound(err) {
			_, stdErr, err := shell.Execute(shell.Command{
				Name: "clusterctl",
				Args: []string{
					"init",
					"--kubeconfig",
					kubeconfigPath,
					"--infrastructure",
					b.provider,
				},
				Tee: true,
			})
			if err != nil {
				return fmt.Errorf("%q: %s", err, stdErr)
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
			err = waitForDeployments(ctx, k8sClient, deploymentKeys)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	return nil
}
