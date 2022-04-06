package kind

import (
	"fmt"
	"os"
	"strings"

	"github.com/giantswarm/capi-bootstrap/pkg/shell"
	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

func (c *Client) CreateCluster(name string, kubeconfigPath string) ([]byte, error) {
	_, stdErr, err := shell.Execute(shell.Command{
		Name: "kind",
		Args: []string{"create", "cluster", "--name", name, "--wait", "2m"},
		Env: map[string]string{
			"KUBECONFIG": kubeconfigPath,
			"PATH":       os.Getenv("PATH"),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%q: %s", err, stdErr)
	}

	return c.GetKubeconfig(name)
}

func (c *Client) DeleteCluster(name string) error {
	_, stdErr, err := shell.Execute(shell.Command{
		Name: "kind",
		Args: []string{"delete", "cluster", "--name", name},
		Env: map[string]string{
			"PATH": os.Getenv("PATH"),
		},
	})
	if err != nil {
		return fmt.Errorf("%q: %s", err, stdErr)
	}

	return nil
}

func (c *Client) GetKubeconfig(name string) ([]byte, error) {
	stdOut, stdErr, err := shell.Execute(shell.Command{
		Name: "kind",
		Args: []string{"get", "kubeconfig", "--name", name},
	})
	if err != nil {
		return nil, fmt.Errorf("%q: %s", err, stdErr)
	}

	return []byte(stdOut), nil
}

func (c *Client) ClusterExists(name string) (bool, error) {
	clusters, err := c.ListClusters()
	if err != nil {
		return false, err
	}

	return util.Contains(clusters, name), nil
}

func (c *Client) ListClusters() ([]string, error) {
	stdOut, stdErr, err := shell.Execute(shell.Command{
		Name: "kind",
		Args: []string{"get", "clusters"},
	})
	if err != nil {
		return nil, fmt.Errorf("%q: %s", err, stdErr)
	}

	return strings.Split(stdOut, "\n"), nil
}
