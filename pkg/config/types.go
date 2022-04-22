package config

import (
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Environment struct {
	Logger     micrologger.Logger
	InCluster  bool
	Type       string
	ConfigFile ConfigFile
	Secrets    map[string]string
}

type ConfigFile struct {
	APIVersion string              `json:"apiVersion"`
	Kind       string              `json:"kind"`
	Spec       BootstrapConfigSpec `json:"spec"`
}

type ClusterConfig struct {
	Kubeconfig string `json:"kubeconfig"`
	Name       string `json:"name"`
}

type AppCollectionRepo struct {
	BranchName string `json:"branchName"`
}

type ConfigRepo struct {
	BranchName string `json:"branchName"`
}

type InstallationsRepo struct {
	BranchName string `json:"branchName"`
}

type BootstrapConfigSpecConfig struct {
	AppCollection AppCollectionRepo `json:"appCollection"`
	Config        ConfigRepo        `json:"config"`
	Installations InstallationsRepo `json:"installations"`
}

type Kubeconfig struct {
	Group string `json:"group"`
	Name  string `json:"name"`
}

type LastpassSecretRef struct {
	Share string `json:"share"`
	Group string `json:"group"`
	Name  string `json:"name"`
}

type EnvironmentVariableRef struct {
	Name string `json:"name"`
}

type Secret struct {
	EnvVar   *EnvironmentVariableRef `json:"envVar,omitempty"`
	Lastpass *LastpassSecretRef      `json:"lastpass,omitempty"`
	Key      string                  `json:"key"`
}

type BootstrapConfigSpec struct {
	BaseDomain       string                    `json:"baseDomain"`
	ClusterNamespace string                    `json:"clusterNamespace"`
	FileInputs       []client.Object           `json:"fileInputs"`
	BootstrapCluster ClusterConfig             `json:"bootstrapCluster"`
	PermanentCluster ClusterConfig             `json:"permanentCluster"`
	Config           BootstrapConfigSpecConfig `json:"config"`
	Kubeconfig       Kubeconfig                `json:"kubeconfig"`
	Secrets          []Secret                  `json:"secrets"`
	Provider         string                    `json:"provider"`
}
