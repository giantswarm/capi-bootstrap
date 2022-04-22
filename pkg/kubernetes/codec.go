package kubernetes

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"github.com/giantswarm/microerror"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
