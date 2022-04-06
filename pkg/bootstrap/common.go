package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v43/github"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/shell"
	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

func (b *Bootstrapper) loadPermanentKubeconfigFromLastPass() error {
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

	b.permanent.KubeconfigPath = kubeconfigFile.Name()
	err = os.WriteFile(b.permanent.KubeconfigPath, data, 0644)
	if err != nil {
		return err
	}

	b.permanent.K8sClient, err = util.KubeconfigToClient(data)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bootstrapper) loadPermanentKubeconfigFromBootstrap(ctx context.Context) error {
	var secret core.Secret
	err := b.bootstrap.K8sClient.Get(ctx, client.ObjectKey{
		Name:      fmt.Sprintf("%s-kubeconfig", b.managementClusterName),
		Namespace: b.clusterNamespace,
	}, &secret)
	if err != nil {
		return err
	}

	return b.loadPermanentKubeconfig(secret.Data["value"])
}

func (b *Bootstrapper) moveCluster(ctx context.Context, bootstrapToPermanent bool) error {
	var source ClusterScope
	var target ClusterScope

	if bootstrapToPermanent {
		source = b.bootstrap
		target = b.permanent
	} else {
		source = b.permanent
		target = b.bootstrap
	}

	// 0. clusterctl move from source to target
	_, stdErr, err := shell.Execute(shell.Command{
		Name: "clusterctl",
		Args: []string{
			"move",
			"--namespace",
			b.clusterNamespace,
			"--kubeconfig",
			source.KubeconfigPath,
			"--to-kubeconfig",
			target.KubeconfigPath,
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
	err = applyResources(ctx, target.K8sClient, clusterResources)
	if err != nil {
		return err
	}

	// 3. wait for apps to become ready
	err = waitForClusterReady(ctx, target.K8sClient, client.ObjectKey{
		Name:      b.managementClusterName,
		Namespace: b.clusterNamespace,
	})
	if err != nil {
		return err
	}

	// 4. delete apps from source cluster
	err = deleteResources(ctx, source.K8sClient, clusterResources)
	if err != nil {
		return err
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
		b.bootstrap.KubeconfigPath,
		b.permanent.KubeconfigPath,
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

func (b *Bootstrapper) getClient(permanent bool) client.Client {
	if permanent {
		return b.permanent.K8sClient
	}
	return b.bootstrap.K8sClient
}

func (b *Bootstrapper) getKubeconfigPath(permanent bool) string {
	if permanent {
		return b.permanent.KubeconfigPath
	}
	return b.bootstrap.KubeconfigPath
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
