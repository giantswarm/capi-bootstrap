package opsctl

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/giantswarm/capi-bootstrap/pkg/shell"
	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

func New(config Config) (Client, error) {
	return Client{
		gitHubToken: config.GitHubToken,
	}, nil
}

func (c *Client) ListInstallations(branch string) ([]string, error) {
	// lpass logout --force
	stdOut, stdErr, err := shell.Execute(shell.Command{
		Name: "opsctl",
		Args: []string{"list", "installations", "--short", "--installations-branch", branch},
		Env: map[string]string{
			"PATH":                        os.Getenv("PATH"),
			"OPSCTL_GITHUB_TOKEN":         c.gitHubToken,
			"OPSCTL_UNSAFE_FORCE_VERSION": "2.7.0",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%q: %s", err, stdErr)
	} else if stdOut == "" {
		return nil, errors.New("no output")
	}

	return strings.Split(strings.TrimSpace(stdOut), " "), nil
}

func (c *Client) InstallationExists(name, branch string) (bool, error) {
	installations, err := c.ListInstallations(branch)
	if err != nil {
		return false, err
	}

	return util.Contains(installations, name), nil
}
