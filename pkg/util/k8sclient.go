package util

import (
	application "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/micrologger"
	core "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func KubeconfigToClient(kubeconfigData []byte) (client.Client, error) {
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigData)
	if err != nil {
		return nil, err
	}

	// TODO: pass this in
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		return nil, err
	}

	k8sClient, err := k8sclient.NewClients(k8sclient.ClientsConfig{
		SchemeBuilder: k8sclient.SchemeBuilder{
			apiextensions.AddToScheme,
			application.AddToScheme,
			core.AddToScheme,
		},
		Logger:     logger,
		RestConfig: restConfig,
	})
	if err != nil {
		return nil, err
	}

	return k8sClient.CtrlClient(), nil
}
