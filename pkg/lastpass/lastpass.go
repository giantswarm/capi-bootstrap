package lastpass

import (
	"os"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/capi-bootstrap/pkg/lastpass/internal/cli"
	"github.com/giantswarm/capi-bootstrap/pkg/lastpass/internal/web"
)

func New() (*Client, error) {
	var client internalClient
	if _, ok := os.LookupEnv("LASTPASS_PASSWORD"); ok {
		var err error
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
		client, err = web.New(web.Config{
			Username:   username,
			Password:   password,
			TOTPSecret: totpSecret,
		})
		if err != nil {
			return nil, microerror.Mask(err)
		}
	} else {
		var err error
		client, err = cli.New()
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return &Client{
		internalClient: client,
	}, nil
}

func mustLookupEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", microerror.Maskf(invalidConfigError, "%s must be defined", key)
	}
	return value, nil
}
