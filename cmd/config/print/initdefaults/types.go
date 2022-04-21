package initdefaults

import (
	"io"

	"github.com/giantswarm/micrologger"
)

type Runner struct {
	flags *Flags

	logger micrologger.Logger

	stdout io.Writer
	stderr io.Writer
}

type Config struct {
	Logger micrologger.Logger

	Stderr io.Writer
	Stdout io.Writer
}

type BootstrapConfig struct {
	APIVersion string              `json:"apiVersion"`
	Kind       string              `json:"kind"`
	Spec       BootstrapConfigSpec `json:"spec"`
}

type BootstrapCluster struct {
	Name string `json:"name"`
}
type PermanentCluster struct {
	Name string `json:"name"`
}

type AppCollection struct {
	BranchName string `json:"branchName"`
}

type Installations struct {
	BranchName string `json:"branchName"`
}

type BootstrapConfigSpecConfig struct {
	AppCollection AppCollection `json:"appCollection"`
	Installations Installations `json:"installations"`
}

type Kubeconfig struct {
	Group string `json:"group"`
	Name  string `json:"name"`
}

type LastpassSecret struct {
	Group           string `json:"group"`
	Name            string `json:"name"`
	SecretName      string `json:"secretName"`
	SecretNamespace string `json:"secretNamespace"`
}

type BootstrapConfigSpec struct {
	BootstrapCluster BootstrapCluster          `json:"bootstrapCluster"`
	PermanentCluster PermanentCluster          `json:"permanentCluster"`
	Config           BootstrapConfigSpecConfig `json:"config"`
	Kubeconfig       Kubeconfig                `json:"kubeconfig"`
	LastpassSecrets  []LastpassSecret          `json:"lastpassSecrets"`
	Provider         string                    `json:"provider"`
}
