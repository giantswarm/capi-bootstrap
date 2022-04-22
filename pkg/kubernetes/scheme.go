package kubernetes

import (
	application "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	builder := runtime.NewSchemeBuilder(
		apiextensions.AddToScheme,
		application.AddToScheme,
		apps.AddToScheme,
		core.AddToScheme)
	err := builder.AddToScheme(scheme)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return scheme, nil
}
