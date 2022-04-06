package main

import (
	"errors"
	"log"
	"os"

	"github.com/google/go-github/v43/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/giantswarm/capi-bootstrap/pkg/bootstrap"
	"github.com/giantswarm/capi-bootstrap/pkg/flags"
	"github.com/giantswarm/capi-bootstrap/pkg/kind"
	"github.com/giantswarm/capi-bootstrap/pkg/lastpass"
	"github.com/giantswarm/capi-bootstrap/pkg/opsctl"
)

func mainE() error {
	ctx := context.Background()

	f := loadFlags()
	err := f.Validate()
	if err != nil {
		log.Fatalln(err.Error())
	}

	opsctlClient, err := opsctl.New(opsctl.Config{
		GitHubToken: os.Getenv("GITHUB_TOKEN"),
	})
	if err != nil {
		return err
	}

	var lastpassClient lastpass.Client
	{
		err = lastpassClient.Login(lastpass.Credentials{
			Username:   f.LastPassUsername,
			Password:   f.LastPassPassword,
			TOTPSecret: f.LastPassTOTPSecret,
		})
		if err != nil {
			return err
		}
	}

	var gitHubClient *github.Client
	{
		token := oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")}
		tokenSource := oauth2.StaticTokenSource(&token)
		httpClient := oauth2.NewClient(ctx, tokenSource)
		gitHubClient = github.NewClient(httpClient)
	}

	var bootstrapper bootstrap.Bootstrapper
	{
		pipeline := "stable"
		if !f.Production {
			pipeline = "testing"
		}
		bootstrapper, err = bootstrap.New(bootstrap.Config{
			ClusterNamespace:      f.ClusterNamespace,
			AccountEngineer:       f.TeamName,
			BaseDomain:            f.CustomerBaseDomain,
			Customer:              f.Customer,
			Pipeline:              pipeline,
			KindClusterName:       f.KindClusterName,
			ManagementClusterName: f.ManagementClusterName,
			Provider:              f.Provider,
			TeamName:              f.TeamName,

			FileInputs: f.Input,

			GitHubClient:   gitHubClient,
			KindClient:     kind.Client{},
			LastPassClient: lastpassClient,
			OpsctlClient:   opsctlClient,
		})
		if err != nil {
			return err
		}
	}

	if f.Command == "create" {
		return bootstrapper.Create(ctx)
	} else if f.Command == "delete" {
		return bootstrapper.Delete(ctx)
	}

	return errors.New("unexpected command")
}

func main() {
	err := mainE()
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func loadFlags() flags.Flags {
	return flags.Flags{
		Command: "delete",

		CustomerBaseDomain:    "test.gigantic.io",
		Customer:              "giantswarm",
		ManagementClusterName: "guppy",
		ClusterNamespace:      "org-giantswarm",
		KindClusterName:       "guppy-bootstrap",
		Provider:              "openstack",
		TeamName:              "Team Rocket",

		Production: false,

		Input: "/tmp/cluster.yaml",
	}
}
