package bootstrap

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"
	"time"

	application "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/google/go-github/v43/github"
	v1 "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apiyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/capi-bootstrap/pkg/shell"
)

var crdGroupVersionKind = schema.GroupVersionKind{
	Group:   "apiextensions.k8s.io",
	Version: "v1",
	Kind:    "CustomResourceDefinition",
}

func New(config Config) (Bootstrapper, error) {
	return Bootstrapper{
		clusterNamespace:      config.ClusterNamespace,
		accountEngineer:       config.AccountEngineer,
		baseDomain:            config.BaseDomain,
		customer:              config.Customer,
		pipeline:              config.Pipeline,
		kindClusterName:       config.KindClusterName,
		managementClusterName: config.ManagementClusterName,
		provider:              config.Provider,
		teamName:              config.TeamName,

		fileInputs: config.FileInputs,

		gitHubClient:   config.GitHubClient,
		kindClient:     config.KindClient,
		lastPassClient: config.LastPassClient,
		opsctlClient:   config.OpsctlClient,
	}, nil
}

func (b *Bootstrapper) getClient(permanent bool) client.Client {
	if permanent {
		return b.permanentK8sClient
	}
	return b.bootstrapK8sClient
}

func (b *Bootstrapper) getKubeconfigPath(permanent bool) string {
	if permanent {
		return b.permanentKubeconfigPath
	}
	return b.bootstrapKubeconfigPath
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

func decodeCRDs(readCloser io.ReadCloser) ([]*apiextensions.CustomResourceDefinition, error) {
	reader := apiyaml.NewYAMLReader(bufio.NewReader(readCloser))
	decoder := scheme.Codecs.UniversalDecoder()

	defer func(contentReader io.ReadCloser) {
		err := readCloser.Close()
		if err != nil {
			panic(err)
		}
	}(readCloser)

	var crds []*apiextensions.CustomResourceDefinition

	for {
		doc, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		//  Skip over empty documents, i.e. a leading `---`
		if len(bytes.TrimSpace(doc)) == 0 {
			continue
		}

		var object unstructured.Unstructured
		_, decodedGVK, err := decoder.Decode(doc, nil, &object)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		switch *decodedGVK {
		case crdGroupVersionKind:
			var crd apiextensions.CustomResourceDefinition
			_, _, err = decoder.Decode(doc, nil, &crd)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			crds = append(crds, &crd)
		default:
			continue
		}
	}

	return crds, nil
}

func decodeObjects(readCloser io.ReadCloser) ([]client.Object, error) {
	reader := apiyaml.NewYAMLReader(bufio.NewReader(readCloser))
	decoder := scheme.Codecs.UniversalDecoder()

	defer func(contentReader io.ReadCloser) {
		err := readCloser.Close()
		if err != nil {
			panic(err)
		}
	}(readCloser)

	var objects []client.Object

	for {
		doc, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		//  Skip over empty documents, i.e. a leading `---`
		if len(bytes.TrimSpace(doc)) == 0 {
			continue
		}

		var object unstructured.Unstructured
		_, _, err = decoder.Decode(doc, nil, &object)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		objects = append(objects, &object)
	}

	return objects, nil
}

func (b *Bootstrapper) fetchAppPlatformCRDs(ctx context.Context) ([]client.Object, error) {
	owner := "giantswarm"
	repo := "apiextensions-application"

	latestRelease, _, err := b.gitHubClient.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	getOptions := github.RepositoryContentGetOptions{
		Ref: "refs/tags/" + *latestRelease.TagName,
	}

	crdPath := path.Join("config", "crd")
	_, contents, _, err := b.gitHubClient.Repositories.GetContents(ctx, "giantswarm", "apiextensions-application", crdPath, &getOptions)
	if err != nil {
		return nil, err
	}

	var crds []client.Object
	for _, file := range contents {
		if filepath.Ext(*file.Name) != ".yaml" {
			continue
		}

		filePath := path.Join(crdPath, *file.Name)
		contentReader, _, err := b.gitHubClient.Repositories.DownloadContents(ctx, "giantswarm", "apiextensions-application", filePath, &getOptions)
		if err != nil {
			return nil, err
		}

		entryCRDs, err := decodeCRDs(contentReader)
		if err != nil {
			return nil, err
		}

		for _, crd := range entryCRDs {
			crds = append(crds, crd)
		}
	}

	return crds, nil
}

func createNamespace(ctx context.Context, k8sClient client.Client, name string) error {
	err := applyResources(ctx, k8sClient, []client.Object{
		&core.Namespace{
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
		},
	})
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

func waitForNamespaces(ctx context.Context, k8sClient client.Client, names []string) error {
	for _, name := range names {
		var ready bool
		for i := 0; i < 300; i++ {
			var namespace core.Namespace
			err := k8sClient.Get(ctx, client.ObjectKey{Name: name}, &namespace)
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return err
			} else if namespace.Status.Phase == "Active" {
				if i > 0 {
					fmt.Print("\n")
				}
				ready = true
				break
			}

			if i == 0 {
				fmt.Printf("waiting for namespace %s to be active", name)
			} else {
				fmt.Print(".")
			}
			time.Sleep(time.Second)
		}

		if !ready {
			return errors.New(fmt.Sprintf("timeout waiting for %s to become active", name))
		}
	}

	return nil
}

func waitForDeployments(ctx context.Context, k8sClient client.Client, keys []client.ObjectKey) error {
	{
		var namespaces []string
		namespaceSet := map[string]struct{}{}
		for _, key := range keys {
			if _, ok := namespaceSet[key.Namespace]; !ok {
				namespaces = append(namespaces, key.Namespace)
				namespaceSet[key.Namespace] = struct{}{}
			}
		}

		err := waitForNamespaces(ctx, k8sClient, namespaces)
		if err != nil {
			return err
		}
	}

	for _, key := range keys {
		var ready bool
		for i := 0; i < 300; i++ {
			var deployment v1.Deployment
			err := k8sClient.Get(ctx, key, &deployment)
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return err
			} else if deployment.Status.Replicas == deployment.Status.ReadyReplicas {
				if i > 0 {
					fmt.Print("\n")
				}
				ready = true
				break
			}

			if i == 0 {
				fmt.Printf("waiting for deployment %s/%s to be ready", key.Namespace, key.Name)
			} else {
				fmt.Print(".")
			}
			time.Sleep(time.Second)
		}

		if !ready {
			return errors.New(fmt.Sprintf("timeout waiting for %s/%s to become ready", key.Namespace, key.Name))
		}
	}

	return nil
}

func waitForCRDs(ctx context.Context, k8sClient client.Client, names []string) error {
	for _, name := range names {
		var ready bool
		for i := 0; i < 300; i++ {
			var crd apiextensions.CustomResourceDefinition
			err := k8sClient.Get(ctx, client.ObjectKey{Name: name}, &crd)
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return err
			} else if crd.Status.AcceptedNames.Kind != "" {
				if i > 0 {
					fmt.Print("\n")
				}
				ready = true
				break
			}

			if i == 0 {
				fmt.Printf("waiting for CRD %s to be created", name)
			} else {
				fmt.Print(".")
			}
			time.Sleep(time.Second)
		}

		if !ready {
			return errors.New(fmt.Sprintf("timeout waiting for CRD %s to be created", name))
		}
	}

	return nil
}

func applyResources(ctx context.Context, k8sClient client.Client, objects []client.Object) error {
	for _, object := range objects {
		err := k8sClient.Create(ctx, object)
		if apierrors.IsAlreadyExists(err) {
			// TODO: possibly handle updates here
			// fall through
		} else if err != nil {
			return err
		}
	}

	return nil
}

func deleteResources(ctx context.Context, k8sClient client.Client, objects []client.Object) error {
	for _, object := range objects {
		err := k8sClient.Delete(ctx, object)
		if apierrors.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return err
		}
	}

	return nil
}
