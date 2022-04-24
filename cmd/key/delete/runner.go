package delete

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/pkg/lastpass"
	"github.com/giantswarm/capi-bootstrap/pkg/sops"
)

func (r *Runner) Run(cmd *cobra.Command, args []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context(), cmd, args)
	return microerror.Mask(err)
}

func (r *Runner) Do(ctx context.Context, _ *cobra.Command, _ []string) error {
	lastpassClient, err := lastpass.New()
	if err != nil {
		return microerror.Mask(err)
	}

	sopsClient, err := sops.New(sops.Config{
		LastpassClient: lastpassClient,
		ClusterName:    r.flag.ClusterName,
	})
	err = sopsClient.DeleteEncryptionKey(ctx)
	return microerror.Mask(err)
}
