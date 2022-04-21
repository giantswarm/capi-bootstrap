package config

import "sigs.k8s.io/controller-runtime/pkg/client"

type BootstrapConfig struct {
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

type LastpassSecrets struct {
	Group           string `json:"group"`
	Name            string `json:"name"`
	SecretName      string `json:"secretName"`
	SecretNamespace string `json:"secretNamespace"`
}

type BootstrapConfigSpec struct {
	ClusterNamespace  string                    `json:"clusterNamespace"`
	FileInputs        []client.Object           `json:"fileInputs"`
	BootstrapCluster  ClusterConfig             `json:"bootstrapCluster"`
	PermanentCluster  ClusterConfig             `json:"permanentCluster"`
	Config            BootstrapConfigSpecConfig `json:"config"`
	Kubeconfig        Kubeconfig                `json:"kubeconfig"`
	LastpassSecrets   []LastpassSecrets         `json:"lastpassSecrets"`
	Provider          string                    `json:"provider"`
}
