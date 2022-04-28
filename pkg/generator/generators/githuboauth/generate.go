package githuboauth

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/capi-bootstrap/pkg/generator/config"
	"github.com/giantswarm/capi-bootstrap/pkg/generator/secret"
)

const Name = "githuboauth"

func New(config config.Config) (*Generator, error) {
	return &Generator{
		client: config.LastpassClient,
	}, nil
}

func (l Generator) Generate(ctx context.Context, secret secret.GeneratedSecretDefinition) (interface{}, error) {
	fmt.Println("Please visit https://github.com/organizations/giantswarm/settings/applications/new to set up a new OAuth app")
	fmt.Printf("Set 'Application name' to: %s-dex\n", secret.ClusterName)
	fmt.Printf("Set 'Homepage URL' to: https://dex.%s.%s\n", secret.ClusterName, secret.BaseDomain)
	fmt.Printf("Set 'Application description' to: %s dex OIDC app\n", secret.ClusterName)
	fmt.Printf("Set 'Authorization callback URL' to: https://dex.%s.%s/callback\n", secret.ClusterName, secret.BaseDomain)
	fmt.Println("Leave 'Enable Device Flow' disabled")
	fmt.Println("Click 'Register application'")

	fmt.Println("Copy the client ID under the 'Client ID' heading and paste it below")
	fmt.Print("Client ID: ")
	var clientID string
	_, err := fmt.Scanf("%s", &clientID)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	fmt.Println("\nClick 'Generate secret' under 'Client secrets', copy the generated secret displayed in the green box, and paste it below")
	fmt.Print("Client secret: ")
	var clientSecret string
	_, err = fmt.Scanf("%s", &clientSecret)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	fmt.Println("\nOperation complete. You may now close the GitHub window.")

	return map[string]string{
		"clientID":     clientID,
		"clientSecret": clientSecret,
	}, nil
}
