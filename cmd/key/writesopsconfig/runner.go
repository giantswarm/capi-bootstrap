package writesopsconfig

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/pkg/sops"
)

func (r *Runner) Run(cmd *cobra.Command, args []string) error {
	err := r.Do(cmd.Context(), cmd, args)
	return microerror.Mask(err)
}

func (r *Runner) Do(ctx context.Context, _ *cobra.Command, _ []string) error {
	sopsClient, err := sops.New(sops.Config{})
	if err != nil {
		return microerror.Mask(err)
	}

	sopsConfig, err := sopsClient.RenderConfig(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = r.stdout.Write(sopsConfig)
	return microerror.Mask(err)
}
