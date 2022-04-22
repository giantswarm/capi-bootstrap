package kubernetes

import (
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	Logger micrologger.Logger

	InCluster  bool
	Kubeconfig string
}

type Client struct {
	logger micrologger.Logger

	ctrlClient  client.Client
	typedClient kubernetes.Interface
}
