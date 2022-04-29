package taylorbot

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/capi-bootstrap/pkg/generator/config"
	"github.com/giantswarm/capi-bootstrap/pkg/templates"
)

const Name = "taylorbot"

func New(config config.Config) (*Generator, error) {
	return &Generator{
		client: config.LastpassClient,
	}, nil
}

func (l Generator) Generate(ctx context.Context, secret templates.TemplateSecret, installation templates.InstallationInputs) (interface{}, error) {
	secretRef := secret.Taylorbot.GitHubCredentialsSecretRef
	credentials, err := l.client.Get(ctx, secretRef.Share, secretRef.Group, secretRef.Name)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	fmt.Println("Please visit https://github.com/settings/tokens/new and log in using the following credentials (log out if already logged in as a different user):")
	fmt.Println("Username:", credentials.Username)
	fmt.Println("Password:", credentials.Password)
	fmt.Println("After logging in:")
	fmt.Println("Set 'Note' to ", installation.ClusterName)
	fmt.Println("Set 'Expiration' to 'No expiration'")
	fmt.Println("Enable all 'repo' scopes (repo:status, repo_deployment, public_repo, repo:invite, security_events)")
	fmt.Println("Copy the generated token from the next page (below 'Make sure to copy your personal access token now. You won’t be able to see it again!')")

	var token string
	fmt.Print("Token secret: ")
	_, err = fmt.Scanf("%s", &token)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	fmt.Println("\nOperation complete. Please log out of the taylorbot account now.")

	return token, nil
}
