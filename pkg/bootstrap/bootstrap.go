package bootstrap

func New(config Config) (Bootstrapper, error) {
	return Bootstrapper{
		clusterNamespace:      config.ClusterNamespace,
		accountEngineer:       config.AccountEngineer,
		baseDomain:            config.BaseDomain,
		customer:              config.Customer,
		pipeline:              config.Pipeline,
		kindClusterName:       config.KindClusterName,
		managementClusterName: config.ManagementClusterName,
		provider:              config.Provider,
		teamName:              config.TeamName,

		fileInputs: config.FileInputs,

		gitHubClient:   config.GitHubClient,
		kindClient:     config.KindClient,
		lastPassClient: config.LastPassClient,
		opsctlClient:   config.OpsctlClient,
	}, nil
}
