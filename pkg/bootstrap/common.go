package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/shell"
	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

func openrcToCloudConfig(content string) CloudConfig {
	cloud := CloudConfigCloud{
		Verify:             false,
		Interface:          "public",
		IdentityAPIVersion: 3,
	}

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "export ") {
			continue
		}

		trimmed = strings.TrimPrefix(trimmed, "export ")
		split := strings.SplitN(trimmed, "=", 2)
		if len(split) != 2 {
			continue
		}

		key := split[0]
		value := strings.Trim(split[1], "\"")

		switch key {
		case "OS_AUTH_URL":
			cloud.Auth.AuthURL = value
		case "OS_PROJECT_ID":
			cloud.Auth.ProjectID = value
		case "OS_USER_DOMAIN_NAME":
			cloud.Auth.UserDomainName = value
		case "OS_USERNAME":
			cloud.Auth.Username = value
		case "OS_PASSWORD":
			cloud.Auth.Password = value
		case "OS_REGION_NAME":
			cloud.RegionName = value
		}
	}

	return CloudConfig{
		Clouds: map[string]CloudConfigCloud{
			"openstack": cloud,
		},
	}
}

func extractClusterCloudConfigName(clusterResources []client.Object) (string, error) {
	for _, resource := range clusterResources {
		asUnstructured, ok := resource.(*unstructured.Unstructured)
		if !ok {
			continue
		} else if asUnstructured.GroupVersionKind().Kind != "ConfigMap" {
			continue
		}

		var configMap core.ConfigMap
		err := runtime.DefaultUnstructuredConverter.
			FromUnstructured(asUnstructured.UnstructuredContent(), &configMap)
		if err != nil {
			return "", err
		}

		if !strings.HasSuffix(configMap.Name, "-cluster-userconfig") {
			continue
		}

		var clusterValues struct {
			CloudConfig string `json:"cloudConfig"`
		}
		err = yaml.Unmarshal([]byte(configMap.Data["values"]), &clusterValues)
		if err != nil {
			return "", err
		}

		return clusterValues.CloudConfig, nil
	}

	return "", errors.New("cluster user values configmap not found")
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

func (b *Bootstrapper) loadPermanentKubeconfigFromLastPass(ctx context.Context) error {
	secretGroup := fmt.Sprintf("Shared-%s/%s\\kubeconfigs", b.teamName, providerShort(b.provider))
	secretName := fmt.Sprintf("%s.kubeconfig", b.managementClusterName)
	kubeconfigSecret, err := b.lastPassClient.GetSecret(secretGroup, secretName)
	if err != nil {
		return err
	}

	err = b.loadPermanentKubeconfig([]byte(kubeconfigSecret.Note))
	if err != nil {
		return err
	}

	return nil
}

func (b *Bootstrapper) loadPermanentKubeconfig(data []byte) error {
	kubeconfigFile, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		return err
	}

	b.permanentKubeconfigPath = kubeconfigFile.Name()
	err = os.WriteFile(b.permanentKubeconfigPath, data, 0644)
	if err != nil {
		return err
	}

	b.permanentK8sClient, err = util.KubeconfigToClient(data)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bootstrapper) loadPermanentKubeconfigFromBootstrap(ctx context.Context) error {
	k8sClient := b.getClient(false)

	var secret core.Secret
	err := k8sClient.Get(ctx, client.ObjectKey{
		Name:      fmt.Sprintf("%s-kubeconfig", b.managementClusterName),
		Namespace: b.clusterNamespace,
	}, &secret)
	if err != nil {
		return err
	}

	return b.loadPermanentKubeconfig(secret.Data["value"])
}

func (b *Bootstrapper) moveCluster(ctx context.Context, bootstrapToPermanent bool) error {
	var sourceK8sClient client.Client
	var sourceKubeconfigPath string
	var targetK8sClient client.Client
	var targetKubeconfigPath string

	if bootstrapToPermanent {
		sourceK8sClient = b.getClient(false)
		sourceKubeconfigPath = b.getKubeconfigPath(false)
		targetK8sClient = b.getClient(true)
		targetKubeconfigPath = b.getKubeconfigPath(true)
	} else {
		sourceK8sClient = b.getClient(true)
		sourceKubeconfigPath = b.getKubeconfigPath(true)
		targetK8sClient = b.getClient(false)
		targetKubeconfigPath = b.getKubeconfigPath(false)
	}

	// 0. clusterctl move from source to target
	_, stdErr, err := shell.Execute(shell.Command{
		Name: "clusterctl",
		Args: []string{
			"move",
			"--namespace",
			b.clusterNamespace,
			"--kubeconfig",
			sourceKubeconfigPath,
			"--to-kubeconfig",
			targetKubeconfigPath,
		},
		Tee: true,
	})
	if err != nil {
		return fmt.Errorf("%q: %s", err, stdErr)
	}

	// 1. read cluster resources from input
	clusterResources, err := decodeObjects(io.NopCloser(strings.NewReader(b.fileInputs)))
	if err != nil {
		return err
	}

	// 2. create cluster apps in target cluster to inherit moved cluster resources
	err = applyResources(ctx, targetK8sClient, clusterResources)
	if err != nil {
		return err
	}

	// 3. wait for apps to become ready
	err = waitForClusterReady(ctx, targetK8sClient, client.ObjectKey{
		Name:      b.managementClusterName,
		Namespace: b.clusterNamespace,
	})
	if err != nil {
		return err
	}

	// 4. delete apps from source cluster
	err = deleteResources(ctx, sourceK8sClient, clusterResources)
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

	b.bootstrapKubeconfigPath = kubeconfigFile.Name()

	var kubeconfigData []byte
	if exists, err := b.kindClient.ClusterExists(b.kindClusterName); err != nil {
		return err
	} else if exists {
		kubeconfigData, err = b.kindClient.GetKubeconfig(b.kindClusterName)
		if err != nil {
			return err
		}

		err = os.WriteFile(b.bootstrapKubeconfigPath, kubeconfigData, 0644)
		if err != nil {
			return err
		}
	} else {
		kubeconfigData, err = b.kindClient.CreateCluster(b.kindClusterName, b.bootstrapKubeconfigPath)
		if err != nil {
			return err
		}
	}

	b.bootstrapK8sClient, err = util.KubeconfigToClient(kubeconfigData)
	if err != nil {
		return err
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

func (b *Bootstrapper) configureHelmCatalogRepo() error {
	_, stdErr, err := shell.Execute(shell.Command{
		Name: "helm",
		Args: []string{
			"repo",
			"add",
			"--force-update",
			"control-plane-catalog",
			"https://giantswarm.github.io/control-plane-catalog",
		},
	})
	if err != nil {
		return fmt.Errorf("%q: %s", err, stdErr)
	}

	_, stdErr, err = shell.Execute(shell.Command{
		Name: "helm",
		Args: []string{
			"repo",
			"update",
		},
	})
	if err != nil {
		return fmt.Errorf("%q: %s", err, stdErr)
	}

	return nil
}

func (b *Bootstrapper) cleanup() {
	for _, file := range []string{
		b.bootstrapKubeconfigPath,
		b.permanentKubeconfigPath,
	} {
		if file == "" {
			continue
		}
		_ = os.Remove(file) // explicitly ignore errors, file might not exist
	}
}

func providerShort(provider string) string {
	switch provider {
	case "aws":
		return "CAPA"
	case "azure":
		return "CAPZ"
	case "gcp":
		return "CAPG"
	case "openstack":
		return "CAPO"
	case "vsphere":
		return "CAPV"
	default:
		return ""
	}
}
