package opsctl

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/giantswarm/capi-bootstrap/pkg/shell"
	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

func New(config Config) (*Client, error) {
	return &Client{
		kubeconfig:            config.Kubeconfig,
		managementClusterName: config.ManagementClusterName,
		gitHubToken:           config.GitHubToken,
		installationsBranch:   config.InstallationsBranch,
	}, nil
}

func (c *Client) EnsureCatalogs() error {
	_, stdErr, err := shell.Execute(shell.Command{
		Name: "opsctl",
		Args: []string{
			"ensure",
			"catalogs",
			"--installation",
			c.managementClusterName,
			"--installations-branch",
			c.installationsBranch,
			"--kubeconfig",
			c.kubeconfig,
		},
		Env: map[string]string{
			"OPSCTL_GITHUB_TOKEN":         c.gitHubToken,
			"PATH":                        os.Getenv("PATH"),
			"HOME":                        os.Getenv("HOME"),
			"OPSCTL_UNSAFE_FORCE_VERSION": "2.10.1-dev", // TODO: read this from `opsctl version` or add a different env var to skip this check completely
		},
	})
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdErr)
	}
	return nil
}

func (c *Client) ListInstallations() ([]string, error) {
	// lpass logout --force
	stdOut, stdErr, err := shell.Execute(shell.Command{
		Name: "opsctl",
		Args: []string{"list", "installations", "--short", "--installations-branch", c.installationsBranch},
		Env: map[string]string{
			"PATH":                        os.Getenv("PATH"),
			"OPSCTL_GITHUB_TOKEN":         c.gitHubToken,
			"OPSCTL_UNSAFE_FORCE_VERSION": "2.7.0",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, stdErr)
	} else if stdOut == "" {
		return nil, errors.New("no output")
	}

	return strings.Split(strings.TrimSpace(stdOut), " "), nil
}

func (c *Client) InstallationExists() (bool, error) {
	installations, err := c.ListInstallations()
	if err != nil {
		return false, err
	}

	return util.Contains(installations, c.managementClusterName), nil
}
