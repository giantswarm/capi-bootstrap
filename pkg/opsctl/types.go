package opsctl

type Config struct {
	ManagementClusterName string
	GitHubToken           string
	InstallationsBranch   string
	Kubeconfig            string
}

type Client struct {
	managementClusterName string
	gitHubToken           string
	installationsBranch   string
	kubeconfig            string
}
