package opsctl

import (
	"fmt"
	"os"

	"github.com/giantswarm/capi-bootstrap/pkg/shell"
)

type Config struct {
	GitHubToken string
}

type Client struct {
	gitHubToken string
}

func (c *Client) EnsureCatalogs(clusterName, installationsBranch, kubeconfigPath string) error {
	_, stdErr, err := shell.Execute(shell.Command{
		Name: "opsctl",
		Args: []string{
			"ensure",
			"catalogs",
			"--installation",
			clusterName,
			"--installations-branch",
			installationsBranch,
			"--kubeconfig",
			kubeconfigPath,
		},
		Env: map[string]string{
			"OPSCTL_GITHUB_TOKEN": c.gitHubToken,
			"PATH":                os.Getenv("PATH"),
			"HOME":                os.Getenv("HOME"),
		},
	})
	if err != nil {
		return fmt.Errorf("%q: %s", err, stdErr)
	}
	return nil
}
