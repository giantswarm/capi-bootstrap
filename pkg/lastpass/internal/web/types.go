package web

import "github.com/ansd/lastpass-go"

type Config struct {
	Username   string
	Password   string
	TOTPSecret string
}

type Client struct {
	client *lastpass.Client

	cachedAccounts []*lastpass.Account

	username   string
	password   string
	totpSecret string
}
