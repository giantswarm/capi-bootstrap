package web

import (
	"context"
	"path/filepath"

	"github.com/ansd/lastpass-go"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/capi-bootstrap/pkg/lastpass/internal"
)

func New(config Config) (*Client, error) {
	return &Client{
		username:   config.Username,
		password:   config.Password,
		totpSecret: config.TOTPSecret,
	}, nil
}

func (c *Client) Create(ctx context.Context, share, group, name, notes string) (*lastpass.Account, error) {
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

func (c *Client) Delete(ctx context.Context, id string) error {
	if err := c.authenticate(ctx); err != nil {
		return microerror.Mask(err)
	}

	c.clearCache()

	err := c.client.Delete(ctx, &lastpass.Account{ID: id})
	return microerror.Mask(err)
}

func (c *Client) Get(ctx context.Context, share, group, name string) (*lastpass.Account, error) {
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

	return nil, microerror.Maskf(internal.NotFoundError, "account %s not found", filepath.Join(share, group, name))
}

func (c *Client) authenticate(ctx context.Context) error {
	if c.client != nil {
		return nil
	}

	totp, err := generateTOTP(c.totpSecret)
	if err != nil {
		return microerror.Mask(err)
	}

	c.client, err = lastpass.NewClient(ctx, c.username, c.password, lastpass.WithOneTimePassword(totp))
	return microerror.Mask(err)
}

func (c *Client) clearCache() {
	c.cachedAccounts = nil
}
