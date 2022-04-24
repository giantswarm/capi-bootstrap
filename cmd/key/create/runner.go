package create

import (
	"context"
	"encoding/json"

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
	if err != nil {
		return microerror.Mask(err)
	}

	encryptionKey, err := sopsClient.EnsureEncryptionKey(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	output, err := json.Marshal(encryptionKey)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = r.stdout.Write(output)
	return microerror.Mask(err)
}
