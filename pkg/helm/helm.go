package helm

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/capi-bootstrap/pkg/shell"
)

var expectedRepos = map[string]string{
	"control-plane-catalog": "https://giantswarm.github.io/control-plane-catalog/",
}

func (c *Client) InstallChart(chartName, catalogName, namespace, values string) error {
	err := c.ensureRepos([]string{catalogName})
	if err != nil {
		return microerror.Mask(err)
	}

	args := []string{
		"upgrade",
		"--install",
		"--namespace",
		namespace,
		"--kubeconfig",
		c.KubeconfigPath,
		chartName,
		fmt.Sprintf("%s/%s", catalogName, chartName),
	}

	if values != "" {
		args = append(args, "--set", values)
	}

	_, stdErr, err := shell.Execute(shell.Command{
		Name: "helm",
		Args: args,
	})
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdErr)
	}

	return nil
}

func (c *Client) ensureRepos(names []string) error {
	var toAdd []string
	var toUpdate []string

	stdOut, stdErr, err := shell.Execute(shell.Command{
		Name: "helm",
		Args: []string{
			"repo",
			"list",
			"--output",
			"json",
		},
	})
	if err != nil {
		return microerror.Mask(fmt.Errorf("%w: %s", err, stdErr))
	}

	var currentRepos []helmRepo
	err = json.Unmarshal([]byte(stdOut), &currentRepos)
	if err != nil {
		return microerror.Mask(err)
	}

	currentReposByName := map[string]string{}
	for _, repo := range currentRepos {
		currentReposByName[repo.Name] = repo.URL
	}

	for _, repoName := range names {
		expectedURL, ok := expectedRepos[repoName]
		if !ok {
			return microerror.Mask(errors.New(fmt.Sprintf("unknown repo: %s", repoName)))
		}

		if currentURL, ok := currentReposByName[repoName]; !ok {
			toAdd = append(toAdd, repoName)
		} else if currentURL != expectedURL {
			toUpdate = append(toUpdate, repoName)
		}
	}

	if len(toAdd) == 0 && len(toUpdate) == 0 {
		return nil
	}

	for _, repoName := range append(toAdd, toUpdate...) {
		url := expectedRepos[repoName]
		_, stdErr, err := shell.Execute(shell.Command{
			Name: "helm",
			Args: []string{
				"repo",
				"add",
				"--force-update",
				repoName,
				url,
			},
		})
		if err != nil {
			return microerror.Mask(fmt.Errorf("%w: %s", err, stdErr))
		}
	}

	_, stdErr, err = shell.Execute(shell.Command{
		Name: "helm",
		Args: []string{
			"repo",
			"update",
		},
	})
	if err != nil {
		return microerror.Mask(fmt.Errorf("%w: %s", err, stdErr))
	}

	return nil
}
