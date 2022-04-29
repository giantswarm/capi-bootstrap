package encrypt

import (
	"context"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

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

	inputFile, err := os.ReadFile(r.flag.InputFile)
	if err != nil {
		return microerror.Mask(err)
	}

	decrypted, err := sopsClient.DecryptSecret(ctx, inputFile)
	if err != nil {
		return microerror.Mask(err)
	}

	decryptedYAML, err := yaml.Marshal(decrypted)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = r.stdout.Write(decryptedYAML)
	return microerror.Mask(err)
}
