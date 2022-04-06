package bootstrap

import (
	"github.com/google/go-github/v43/github"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/capi-bootstrap/pkg/kind"
	"github.com/giantswarm/capi-bootstrap/pkg/lastpass"
	"github.com/giantswarm/capi-bootstrap/pkg/opsctl"
)

type Config struct {
	ClusterNamespace string

	AccountEngineer       string
	BaseDomain            string
	Customer              string
	Pipeline              string
	KindClusterName       string
	ManagementClusterName string
	Provider              string
	TeamName              string

	FileInputs string

	GitHubClient   *github.Client
	KindClient     kind.Client
	LastPassClient lastpass.Client
	OpsctlClient   opsctl.Client
}

type Bootstrapper struct {
	clusterNamespace string

	accountEngineer       string
	baseDomain            string
	customer              string
	pipeline              string
	kindClusterName       string
	managementClusterName string
	provider              string
	teamName              string

	fileInputs string

	gitHubClient   *github.Client
	kindClient     kind.Client
	lastPassClient lastpass.Client
	opsctlClient   opsctl.Client

	bootstrapK8sClient      client.Client
	bootstrapKubeconfigPath string

	permanentK8sClient      client.Client
	permanentKubeconfigPath string
}

type CloudConfigCloudAuth struct {
	AuthURL        string `json:"auth_url"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	UserDomainName string `json:"user_domain_name"`
	ProjectID      string `json:"project_id"`
}

type CloudConfigCloud struct {
	Auth               CloudConfigCloudAuth `json:"auth"`
	Verify             bool                 `json:"verify"`
	RegionName         string               `json:"region_name"`
	Interface          string               `json:"interface"`
	IdentityAPIVersion int                  `json:"identity_api_version"`
}

type CloudConfig struct {
	Clouds map[string]CloudConfigCloud `json:"clouds"`
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
