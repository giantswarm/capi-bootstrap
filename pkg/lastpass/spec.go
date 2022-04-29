package lastpass

import (
	"context"

	"github.com/ansd/lastpass-go"
)

type internalClient interface {
	Create(ctx context.Context, share, group, name, notes string) (*lastpass.Account, error)
	Get(ctx context.Context, share, group, name string) (*lastpass.Account, error)
	Delete(ctx context.Context, id string) error
}
