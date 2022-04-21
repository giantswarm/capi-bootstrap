package repo

import (
	"github.com/google/go-github/v43/github"
)

type Config struct {
	GitHubClient *github.Client

	AccountEngineer       string
	BaseDomain            string
	Customer              string
	ManagementClusterName string
	Pipeline              string
	Provider              string
}

type Service struct {
	gitHubClient *github.Client

	accountEngineer       string
	baseDomain            string
	customer              string
	managementClusterName string
	pipeline              string
	provider              string
}

type ClusterDefinition struct {
	Base            string `json:"base"`
	Codename        string `json:"codename"`
	Customer        string `json:"customer"`
	AccountEngineer string `json:"accountEngineer"`
	Pipeline        string `json:"pipeline"`
	Provider        string `json:"provider"`
}

type AppCatalogValues struct {
	AppCatalog AppCatalog `json:"appCatalog"`
}

type AppCatalog struct {
	Config AppCatalogConfig `json:"config"`
}

type AppCatalogConfig struct {
	ConfigMap AppCatalogConfigConfigMap `json:"configMap"`
}

type AppCatalogConfigConfigMap struct {
	Values AppCatalogConfigConfigMapValues `json:"values"`
}

type AppCatalogConfigConfigMapValues struct {
	BaseDomain        string `json:"baseDomain"`
	ManagementCluster string `json:"managementCluster"`
	Provider          string `json:"provider"`
}
