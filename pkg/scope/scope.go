package scope

import "sigs.k8s.io/controller-runtime/pkg/client"

type Cluster struct {
	KubeconfigPath string
	K8sClient      client.Client
}

type Scope struct {
	Bootstrap *Cluster
	Permanent *Cluster
}
