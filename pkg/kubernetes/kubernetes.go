package kubernetes

import (
	"context"
	"io"
	"time"

	application "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func New(config Config) (*Client, error) {
	var restConfig *rest.Config
	if config.InCluster {
		var err error
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, microerror.Mask(err)
		}
	} else {
		var err error
		restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{
				ExplicitPath: config.Kubeconfig,
			}, nil).ClientConfig()
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	scheme, err := NewScheme()
	if err != nil {
		return nil, microerror.Mask(err)
	}
	ctrlClient, err := client.New(restConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	typedClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &Client{
		ctrlClient:  ctrlClient,
		typedClient: typedClient,

		logger: config.Logger,
	}, nil
}

func (c Client) WaitForNamespaces(ctx context.Context, names []string) error {
	for _, name := range names {
		var ready bool
		for i := 0; !ready; i++ {
			var namespace core.Namespace
			err := c.ctrlClient.Get(ctx, client.ObjectKey{Name: name}, &namespace)
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			} else if namespace.Status.Phase == "Active" {
				if i > 0 {
					c.logger.Debugf(ctx, "namespace %s is ready", name)
				}
				ready = true
				break
			}

			if i == 0 {
				c.logger.Debugf(ctx, "waiting for namespace %s to be ready", name)
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
			err := c.ctrlClient.Get(ctx, key, &deployment)
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			} else if deployment.Status.Replicas == deployment.Status.ReadyReplicas {
				c.logger.Debugf(ctx, "deployment %s/%s is ready", key.Namespace, key.Name)
				ready = true
				break
			}

			if i == 0 {
				c.logger.Debugf(ctx, "waiting for deployment %s/%s to be ready", key.Namespace, key.Name)
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
			err := c.ctrlClient.Get(ctx, client.ObjectKey{Name: name}, &crd)
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			} else if crd.Status.AcceptedNames.Kind != "" {
				c.logger.Debugf(ctx, "CRD %s found", name)
				ready = true
				break
			}

			if i == 0 {
				c.logger.Debugf(ctx, "waiting for CRD %s to be created", name)
			}

			time.Sleep(time.Second * 10)
		}
	}

	return nil
}

func (c Client) ApplyResources(ctx context.Context, objects []client.Object) error {
	for _, object := range objects {
		err := c.ctrlClient.Create(ctx, object)
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
		err := c.ctrlClient.Delete(ctx, object)
		if apierrors.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
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
		err := c.ctrlClient.List(ctx, &namespaceApps, client.InNamespace(namespace))
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
				c.logger.Debugf(ctx, "waiting for app %s to be created", appName)
				allAppsDeployed = false
				break
			} else if app.Status.Release.Status != "deployed" {
				c.logger.Debugf(ctx, "waiting for app %s to have status \"deployed\", current status \"%s\"", appName, app.Status.Release.Status)
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

func (c Client) WatchPodLogs(ctx context.Context, appName string) error {
	var podList core.PodList
	err := c.ctrlClient.List(ctx, &podList, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": "capi-bootstrap",
		}),
		Namespace: "giantswarm",
		Limit:     1,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	c.logger.Debugf(ctx, "found capi-bootstrap pod %s", podList.Items[0].Name)

	request := c.typedClient.CoreV1().Pods("giantswarm").GetLogs(podList.Items[0].Name, &core.PodLogOptions{
		Container:  "capi-bootstrap",
		Follow:     true,
		LimitBytes: nil,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	podLogs, err := request.Stream(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	defer podLogs.Close()

	c.logger.Debugf(ctx, "following pod logs")

	for {
		buffer := make([]byte, 100)
		bytesRead, err := podLogs.Read(buffer)
		if bytesRead > 0 {
			c.logger.Debugf(ctx, string(buffer[:bytesRead]))
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
