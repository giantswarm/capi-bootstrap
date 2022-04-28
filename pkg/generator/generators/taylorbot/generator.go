package taylorbot

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/capi-bootstrap/pkg/generator/config"
	"github.com/giantswarm/capi-bootstrap/pkg/generator/secret"
)

const Name = "taylorbot"

func New(config config.Config) (*Generator, error) {
	return &Generator{
		client: config.LastpassClient,
	}, nil
}

func (l Generator) Generate(ctx context.Context, secret secret.GeneratedSecretDefinition) (interface{}, error) {
	secretRef := secret.Taylorbot.GitHubCredentialsSecretRef
	credentials, err := l.client.GetAccount(ctx, secretRef.Share, secretRef.Group, secretRef.Name)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	fmt.Println("Please visit https://github.com/settings/tokens/new and log in using the following credentials:")
	fmt.Println("Username:", credentials.Username)
	fmt.Println("Password:", credentials.Password)
	fmt.Println("After logging in:")
	fmt.Println("Set 'Note' to ", secret.ClusterName)
	fmt.Println("Set 'Expiration' to 'No expiration'")
	fmt.Println("Enable all 'repo' scopes (repo:status, repo_deployment, public_repo, repo:invite, security_events)")
	fmt.Println("Copy the generated token from the next page (below 'Make sure to copy your personal access token now. You won’t be able to see it again!')")
	var token string
	fmt.Print("Token secret: ")
	_, err = fmt.Scanf("%s", &token)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	fmt.Println("Operation complete. Please log out of the taylorbot account now.")

	return token, nil
}
