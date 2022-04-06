package bootstrap

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/giantswarm/microerror"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apiyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var crdGroupVersionKind = schema.GroupVersionKind{
	Group:   "apiextensions.k8s.io",
	Version: "v1",
	Kind:    "CustomResourceDefinition",
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
			var deployment apps.Deployment
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
