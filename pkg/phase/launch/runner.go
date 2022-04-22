package launch

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	config2 "github.com/giantswarm/capi-bootstrap/pkg/config"
)

func (r *Runner) Run(cmd *cobra.Command, _ []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	environment, err := r.flag.BuildEnvironment(r.logger)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context(), environment)
	return microerror.Mask(err)
}

func (r *Runner) Do(ctx context.Context, environment *config2.Environment) error {
	k8sClient, err := environment.GetK8sClient()
	if err != nil {
		return microerror.Mask(err)
	}

	helmClient, err := environment.GetHelmClient()
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "launching capi-bootstrap job")

	err = helmClient.InstallChart("capi-bootstrap", "control-plane-catalog", "giantswarm", "")
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "launched capi-bootstrap job")

	err = k8sClient.WatchPodLogs(ctx, "capi-bootstrap")
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
