package lastpass

import (
	"context"
	"os"
	"path/filepath"

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

	c.clearCache()

	return &toCreate, nil
}

func (c *Client) GetAccount(ctx context.Context, share, group, name string) (*lastpass.Account, error) {
	if err := c.authenticate(ctx); err != nil {
		return nil, microerror.Mask(err)
	}

	if c.cachedAccounts == nil {
		accounts, err := c.client.Accounts(ctx)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		c.cachedAccounts = accounts
	}

	for _, account := range c.cachedAccounts {
		if account.Share == share && account.Name == name && account.Group == group {
			return account, nil
		}
	}

	return nil, microerror.Maskf(notFoundError, "account %s not found", filepath.Join(share, group, name))
}

func (c *Client) DeleteAccount(ctx context.Context, id string) error {
	if err := c.authenticate(ctx); err != nil {
		return microerror.Mask(err)
	}

	c.clearCache()

	err := c.client.Delete(ctx, &lastpass.Account{ID: id})
	return microerror.Mask(err)
}

func (c *Client) clearCache() {
	c.cachedAccounts = nil
}
