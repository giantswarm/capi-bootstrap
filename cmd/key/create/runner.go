package create

import (
	"context"
	"fmt"

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

	output := fmt.Sprintf(`export SOPS_AGE_KEY=%s
export SOPS_AGE_RECIPIENTS=%s
`, encryptionKey.PrivateKey, encryptionKey.PublicKey)

	_, err = r.stdout.Write([]byte(output))
	return microerror.Mask(err)
}
