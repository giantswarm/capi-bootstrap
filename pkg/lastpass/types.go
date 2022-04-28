package lastpass

import (
	"github.com/ansd/lastpass-go"
)

type Client struct {
	client *lastpass.Client

	cachedAccounts []*lastpass.Account

	username   string
	password   string
	totpSecret string
}
