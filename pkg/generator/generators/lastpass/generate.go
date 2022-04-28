package lastpass

import (
	"context"

	"github.com/giantswarm/microerror"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/generator/config"
	"github.com/giantswarm/capi-bootstrap/pkg/generator/secret"
)

const Name = "lastpass"

func New(config config.Config) (*Generator, error) {
	return &Generator{
		client: config.LastpassClient,
	}, nil
}

func (l Generator) Generate(ctx context.Context, secret secret.GeneratedSecretDefinition) (interface{}, error) {
	secretRef := secret.Lastpass.SecretRef
	account, err := l.client.GetAccount(ctx, secretRef.Share, secretRef.Group, secretRef.Name)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if secret.Lastpass.Format == "yaml" {
		var data map[string]string
		err = yaml.Unmarshal([]byte(account.Notes), &data)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		return data, nil
	}

	return account.Notes, nil
}
