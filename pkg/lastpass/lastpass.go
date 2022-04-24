package lastpass

import (
	"context"
	"os"

	"github.com/ansd/lastpass-go"
	"github.com/giantswarm/microerror"
)

func mustLookupEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", microerror.Maskf(invalidConfigError, "%s must be defined", key)
	}
	return value, nil
}

func New() (*Client, error) {
	username, err := mustLookupEnv("LASTPASS_USERNAME")
	if err != nil {
		return nil, microerror.Mask(err)
	}
	password, err := mustLookupEnv("LASTPASS_PASSWORD")
	if err != nil {
		return nil, microerror.Mask(err)
	}
	totpSecret, err := mustLookupEnv("LASTPASS_TOTP_SECRET")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &Client{
		username:   username,
		password:   password,
		totpSecret: totpSecret,
	}, nil
}

func (c *Client) authenticate(ctx context.Context) error {
	if c.client == nil {
		totp, err := generateTOTP(c.totpSecret)
		if err != nil {
			return microerror.Mask(err)
		}

		c.client, err = lastpass.NewClient(ctx, c.username, c.password, lastpass.WithOneTimePassword(totp))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (c *Client) CreateAccount(ctx context.Context, share, group, name, notes string) (*lastpass.Account, error) {
	if err := c.authenticate(ctx); err != nil {
		return nil, microerror.Mask(err)
	}

	toCreate := lastpass.Account{
		Name:  name,
		Group: group,
		Share: share,
		Notes: notes,
	}
	err := c.client.Add(ctx, &toCreate)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &toCreate, nil
}

func (c *Client) GetAccount(ctx context.Context, share, group, name string) (*lastpass.Account, error) {
	if err := c.authenticate(ctx); err != nil {
		return nil, microerror.Mask(err)
	}

	accounts, err := c.client.Accounts(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, account := range accounts {
		if account.Share == share && account.Name == name && account.Group == group {
			return account, nil
		}
	}

	return nil, microerror.Maskf(notFoundError, "account %s/%s/%s not found", share, group, name)
}
