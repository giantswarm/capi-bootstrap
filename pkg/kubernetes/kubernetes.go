package kubernetes

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	application "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apiyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

var crdGroupVersionKind = schema.GroupVersionKind{
	Group:   "apiextensions.k8s.io",
	Version: "v1",
	Kind:    "CustomResourceDefinition",
}

type Client struct {
	client.Client

	Logger micrologger.Logger
}

func ClientFromFlags(kubeconfigPath string, inCluster bool) (*Client, error) {
	var restConfig *rest.Config
	if inCluster {
		var err error
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, microerror.Mask(err)
		}
	} else {
		var err error
		restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{
				ExplicitPath: kubeconfigPath,
			}, nil).ClientConfig()
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	clientScheme := runtime.NewScheme()
	schemeBuilder := runtime.NewSchemeBuilder(
		apiextensions.AddToScheme,
		application.AddToScheme,
		apps.AddToScheme,
		core.AddToScheme)
	err := schemeBuilder.AddToScheme(clientScheme)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sClient, err := client.New(restConfig, client.Options{
		Scheme: clientScheme,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &Client{Client: k8sClient}, nil
}

func TypedClientFromFlags(kubeconfigPath string, inCluster bool) (kubernetes.Interface, error) {
	var restConfig *rest.Config
	if inCluster {
		var err error
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, microerror.Mask(err)
		}
	} else {
		var err error
		restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{
				ExplicitPath: kubeconfigPath,
			}, nil).ClientConfig()
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	clientScheme := runtime.NewScheme()
	schemeBuilder := runtime.NewSchemeBuilder(
		apiextensions.AddToScheme,
		application.AddToScheme,
		apps.AddToScheme,
		core.AddToScheme)
	err := schemeBuilder.AddToScheme(clientScheme)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return k8sClient, nil
}

func (c Client) WaitForNamespaces(ctx context.Context, names []string) error {
	for _, name := range names {
		var ready bool
		for i := 0; !ready; i++ {
			var namespace core.Namespace
			err := c.Get(ctx, client.ObjectKey{Name: name}, &namespace)
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			} else if namespace.Status.Phase == "Active" {
				if i > 0 {
					c.Logger.Debugf(ctx, "namespace %s is ready", name)
				}
				ready = true
				break
			}

			if i == 0 {
				c.Logger.Debugf(ctx, "waiting for namespace %s to be ready", name)
			}

			time.Sleep(time.Second * 10)
		}
	}

	return nil
}

func (c Client) WaitForDeployments(ctx context.Context, keys []client.ObjectKey) error {
	{
		var namespaces []string
		namespaceSet := map[string]struct{}{}
		for _, key := range keys {
			if _, ok := namespaceSet[key.Namespace]; !ok {
				namespaces = append(namespaces, key.Namespace)
				namespaceSet[key.Namespace] = struct{}{}
			}
		}

		err := c.WaitForNamespaces(ctx, namespaces)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	for _, key := range keys {
		var ready bool
		for i := 0; !ready; i++ {
			var deployment apps.Deployment
			err := c.Get(ctx, key, &deployment)
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			} else if deployment.Status.Replicas == deployment.Status.ReadyReplicas {
				c.Logger.Debugf(ctx, "deployment %s/%s is ready", key.Namespace, key.Name)
				ready = true
				break
			}

			if i == 0 {
				c.Logger.Debugf(ctx, "waiting for deployment %s/%s to be ready", key.Namespace, key.Name)
			}

			time.Sleep(time.Second * 10)
		}
	}

	return nil
}

func (c Client) WaitForCRDs(ctx context.Context, names []string) error {
	for _, name := range names {
		var ready bool
		for i := 0; !ready; i++ {
			var crd apiextensions.CustomResourceDefinition
			err := c.Get(ctx, client.ObjectKey{Name: name}, &crd)
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			} else if crd.Status.AcceptedNames.Kind != "" {
				c.Logger.Debugf(ctx, "CRD %s found", name)
				ready = true
				break
			}

			if i == 0 {
				c.Logger.Debugf(ctx, "waiting for CRD %s to be created", name)
			}

			time.Sleep(time.Second * 10)
		}
	}

	return nil
}

func (c Client) ApplyResources(ctx context.Context, objects []client.Object) error {
	for _, object := range objects {
		err := c.Create(ctx, object)
		if apierrors.IsAlreadyExists(err) {
			// TODO: possibly handle updates here
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (c Client) DeleteResources(ctx context.Context, objects []client.Object) error {
	for _, object := range objects {
		err := c.Delete(ctx, object)
		if apierrors.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func DecodeCRDs(readCloser io.ReadCloser) ([]*apiextensions.CustomResourceDefinition, error) {
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

func DecodeObjects(readCloser io.ReadCloser) ([]client.Object, error) {
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

func (c Client) CreateNamespace(ctx context.Context, name string) error {
	err := c.ApplyResources(ctx, []client.Object{
		&core.Namespace{
			ObjectMeta: meta.ObjectMeta{
				Name: name,
			},
		},
	})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (c Client) WaitForAppsDeployed(ctx context.Context, namespace string, names []string) error {
	for {
		var namespaceApps application.AppList
		err := c.List(ctx, &namespaceApps, client.InNamespace(namespace))
		if err != nil {
			return microerror.Mask(err)
		}

		appsByName := map[string]application.App{}
		for _, app := range namespaceApps.Items {
			appsByName[app.ObjectMeta.Name] = app
		}

		allAppsDeployed := true
		for _, appName := range names {
			if app, ok := appsByName[appName]; !ok {
				c.Logger.Debugf(ctx, "waiting for app %s to be created", appName)
				allAppsDeployed = false
				break
			} else if app.Status.Release.Status != "deployed" {
				c.Logger.Debugf(ctx, "waiting for app %s to have status \"deployed\", current status \"%s\"", appName, app.Status.Release.Status)
				allAppsDeployed = false
				break
			}
		}

		if allAppsDeployed {
			break
		}

		time.Sleep(time.Second * 10)
	}

	return nil
}

func (c Client) CreateCloudConfigSecret(ctx context.Context, openrcContent string, clusterNamespace string) error {
	cloudConfig := util.OpenrcToCloudConfig(openrcContent)
	cloudConfigYAML, err := yaml.Marshal(cloudConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	secret := core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Labels: map[string]string{
				"clusterctl.cluster.x-k8s.io/move": "true",
			},
			Name:      "cloud-config",
			Namespace: clusterNamespace,
		},
		StringData: map[string]string{
			"clouds.yaml": string(cloudConfigYAML),
		},
	}

	err = c.ApplyResources(ctx, []client.Object{&secret})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (c Client) WaitForClusterReady(ctx context.Context, clusterNamespace, clusterName string) error {
	var appNames []string
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
		appNames = append(appNames, fmt.Sprintf("%s-%s", clusterName, appName))
	}

	err := c.WaitForAppsDeployed(ctx, clusterNamespace, appNames)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (c Client) WaitForClusterDeleted(ctx context.Context, clusterNamespace, clusterName string) error {
	for {
		var clusterApps application.AppList
		err := c.List(ctx, &clusterApps, client.InNamespace(clusterNamespace), client.MatchingLabels{
			"giantswarm.io/cluster": clusterName,
		})
		if err != nil {
			return microerror.Mask(err)
		}

		if len(clusterApps.Items) == 0 {
			c.Logger.Debugf(ctx, "all apps deleted")
			break
		}

		appsByName := map[string]application.App{}
		for _, app := range clusterApps.Items {
			appName := strings.TrimPrefix(app.ObjectMeta.Name, fmt.Sprintf("%s-", clusterName))
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
				c.Logger.Debugf(ctx, "waiting for app %s to be deleted", appName)
				allAppsDeleted = false
				break
			}
		}

		if allAppsDeleted {
			c.Logger.Debugf(ctx, "all apps deleted")
			break
		}

		time.Sleep(time.Second * 10)
	}

	c.Logger.Debugf(ctx, "waiting for cluster to be deleted")

	for {
		cluster := meta.PartialObjectMetadata{
			TypeMeta: meta.TypeMeta{
				Kind:       "Cluster",
				APIVersion: "cluster.x-k8s.io/v1beta1",
			},
		}
		err := c.Get(ctx, client.ObjectKey{
			Name:      clusterName,
			Namespace: clusterNamespace,
		}, &cluster)
		if apierrors.IsNotFound(err) {
			c.Logger.Debugf(ctx, "cluster deleted")
			break
		} else if err != nil {
			return microerror.Mask(err)
		}

		time.Sleep(time.Second * 10)
	}

	return nil
}
