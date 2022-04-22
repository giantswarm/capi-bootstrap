package config

import (
	"context"
	"os"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/helm"
	"github.com/giantswarm/capi-bootstrap/pkg/key"
	"github.com/giantswarm/capi-bootstrap/pkg/kind"
	"github.com/giantswarm/capi-bootstrap/pkg/kubernetes"
	"github.com/giantswarm/capi-bootstrap/pkg/lastpass"
	"github.com/giantswarm/capi-bootstrap/pkg/opsctl"
	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

func (e *Environment) GetGitHubClient(ctx context.Context) (*github.Client, error) {
	var accessToken string
	if e.InCluster {
		var ok bool
		accessToken, ok = e.Secrets["github-token"]
		if !ok {
			return nil, microerror.Maskf(notFoundError, "github-token secret not found")
		}
	} else {
		var ok bool
		accessToken, ok = os.LookupEnv("GITHUB_TOKEN")
		if !ok {
			return nil, microerror.Maskf(notFoundError, "GITHUB_TOKEN environment variable not found")
		}
	}

	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
	}))

	return github.NewClient(httpClient), nil
}

func (e *Environment) getGitHubToken() (string, error) {
	var accessToken string
	if e.InCluster {
		var ok bool
		accessToken, ok = e.Secrets["github-token"]
		if !ok {
			return "", microerror.Maskf(notFoundError, "github-token secret not found")
		}
	} else {
		var ok bool
		accessToken, ok = os.LookupEnv("GITHUB_TOKEN")
		if !ok {
			return "", microerror.Maskf(notFoundError, "GITHUB_TOKEN environment variable not found")
		}
	}

	return accessToken, nil
}

func (e *Environment) GetKubeconfig() (string, error) {
	if e.InCluster {
		err := generateInClusterKubeconfig("in-cluster.kubeconfig")
		if err != nil {
			return "", microerror.Mask(err)
		}

		return "in-cluster.kubeconfig", nil
	} else if e.Type == "bootstrap" {
		return e.ConfigFile.Spec.BootstrapCluster.Kubeconfig, nil
	} else if e.Type == "permanent" {
		return e.ConfigFile.Spec.PermanentCluster.Kubeconfig, nil
	}

	return "", microerror.Maskf(invalidConfigError, "unknown environment type %s", e.Type)
}

func (e *Environment) GetHelmClient() (*helm.Client, error) {
	kubeconfig, err := e.GetKubeconfig()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &helm.Client{
		KubeconfigPath: kubeconfig,
	}, nil
}

func (e *Environment) GetKindClient() (*kind.Client, error) {
	return &kind.Client{
		ClusterName: e.ConfigFile.Spec.PermanentCluster.Name,
	}, nil
}

func (e *Environment) GetLastpassClient() (*lastpass.Client, error) {
	username, ok := os.LookupEnv("LASTPASS_USERNAME")
	if !ok {
		return nil, microerror.Maskf(invalidConfigError, "LASTPASS_USERNAME must be defined")
	}

	password, ok := os.LookupEnv("LASTPASS_PASSWORD")
	if !ok {
		return nil, microerror.Maskf(invalidConfigError, "LASTPASS_USERNAME must be defined")
	}

	totpSecret, ok := os.LookupEnv("LASTPASS_TOTP_SECRET")
	if !ok {
		return nil, microerror.Maskf(invalidConfigError, "LASTPASS_USERNAME must be defined")
	}

	lastpassClient, err := lastpass.New(lastpass.Config{
		Username:   username,
		Password:   password,
		TOTPSecret: totpSecret,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return lastpassClient, nil
}

func (e *Environment) GetOpsctlClient() (*opsctl.Client, error) {
	kubeconfig, err := e.GetKubeconfig()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	gitHubToken, err := e.getGitHubToken()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	opsctlClient, err := opsctl.New(opsctl.Config{
		ManagementClusterName: e.ConfigFile.Spec.PermanentCluster.Name,
		GitHubToken:           gitHubToken,
		InstallationsBranch:   e.ConfigFile.Spec.Config.Installations.BranchName,
		Kubeconfig:            kubeconfig,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return opsctlClient, nil
}

func (e *Environment) GetK8sClient() (*kubernetes.Client, error) {
	var kubeconfig string
	if !e.InCluster {
		var err error
		kubeconfig, err = e.GetKubeconfig()
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	k8sClient, err := kubernetes.New(kubernetes.Config{
		Logger:     e.Logger,
		InCluster:  e.InCluster,
		Kubeconfig: kubeconfig,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return k8sClient, nil
}

func generateInClusterKubeconfig(path string) error {
	kubeconfig := api.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: map[string]*api.Cluster{
			"default": {
				Server:               "https://kubernetes.default.svc",
				CertificateAuthority: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			"default": {
				TokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
			},
		},
		Contexts: map[string]*api.Context{
			"default": {
				Cluster:   "default",
				AuthInfo:  "default",
				Namespace: "default",
			},
		},
		CurrentContext: "default",
	}

	err := clientcmd.WriteToFile(kubeconfig, path)
	return microerror.Mask(err)
}

func (b ConfigFile) ToFile(file string) error {
	content, err := yaml.Marshal(b)
	if err != nil {
		return microerror.Mask(err)
	}

	err = os.WriteFile(file, content, 0644)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (b ConfigFile) Validate() error {
	if b.Kind != "BootstrapConfig" {
		return microerror.Maskf(invalidFlagError, "invalid kind %s", b.Kind)
	}
	if b.APIVersion != "v1alpha1" {
		return microerror.Maskf(invalidFlagError, "invalid api version %s", b.APIVersion)
	}
	if !util.Contains(key.AllowedProviders, b.Spec.Provider) {
		return microerror.Maskf(invalidFlagError, "provider must be one of: %s", strings.Join(key.AllowedProviders, ","))
	}
	return nil
}
