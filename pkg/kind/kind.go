package kind

import (
	"fmt"
	"os"
	"strings"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/capi-bootstrap/pkg/shell"
	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

func (c *Client) CreateCluster(kubeconfigPath string) ([]byte, error) {
	_, stdErr, err := shell.Execute(shell.Command{
		Name: "kind",
		Args: []string{"create", "cluster", "--name", c.ClusterName, "--wait", "2m"},
		Env: map[string]string{
			"KUBECONFIG": kubeconfigPath,
			"PATH":       os.Getenv("PATH"),
		},
	})
	if err != nil {
		return nil, microerror.Mask(fmt.Errorf("%w: %s", err, stdErr))
	}

	return c.GetKubeconfig()
}

func (c *Client) DeleteCluster() error {
	_, stdErr, err := shell.Execute(shell.Command{
		Name: "kind",
		Args: []string{"delete", "cluster", "--name", c.ClusterName},
		Env: map[string]string{
			"PATH": os.Getenv("PATH"),
		},
	})
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdErr)
	}

	return nil
}

func (c *Client) GetKubeconfig() ([]byte, error) {
	stdOut, stdErr, err := shell.Execute(shell.Command{
		Name: "kind",
		Args: []string{"get", "kubeconfig", "--name", c.ClusterName},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, stdErr)
	}

	return []byte(stdOut), nil
}

func (c *Client) ClusterExists() (bool, error) {
	clusters, err := c.ListClusters()
	if err != nil {
		return false, err
	}

	return util.Contains(clusters, c.ClusterName), nil
}

func (c *Client) ListClusters() ([]string, error) {
	stdOut, stdErr, err := shell.Execute(shell.Command{
		Name: "kind",
		Args: []string{"get", "clusters"},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, stdErr)
	}

	return strings.Split(stdOut, "\n"), nil
}
