package lastpass

import (
	"context"

	"github.com/ansd/lastpass-go"
	"github.com/giantswarm/microerror"
)

func New(config Config) (*Client, error) {
	if config.Username == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Username must not be empty", config)
	}
	if config.Password == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Password must not be empty", config)
	}
	if config.TOTPSecret == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.TOTPSecret must not be empty", config)
	}

	return &Client{
		client: nil,

		username:   config.Username,
		password:   config.Password,
		totpSecret: config.TOTPSecret,
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
