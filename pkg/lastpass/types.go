package lastpass

import (
	"context"

	"github.com/ansd/lastpass-go"
	"github.com/giantswarm/microerror"
)

type Config struct {
	Username   string
	Password   string
	TOTPSecret string
}

type Client struct {
	client *lastpass.Client

	username   string
	password   string
	totpSecret string
}

func (c *Client) DeleteAccount(ctx context.Context, id string) error {
	if err := c.authenticate(ctx); err != nil {
		return microerror.Mask(err)
	}

	err := c.client.Delete(ctx, &lastpass.Account{ID: id})
	return microerror.Mask(err)
}
