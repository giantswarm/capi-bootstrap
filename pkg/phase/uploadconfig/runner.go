package uploadconfig

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	config2 "github.com/giantswarm/capi-bootstrap/pkg/config"
	"github.com/giantswarm/capi-bootstrap/pkg/kubernetes"
)

func (r *Runner) Run(cmd *cobra.Command, _ []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	bootstrapConfig, err := r.flag.ToConfig()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context(), bootstrapConfig)
	return microerror.Mask(err)
}

func (r *Runner) Do(ctx context.Context, bootstrapConfig config2.BootstrapConfig) error {
	k8sClient, err := kubernetes.ClientFromFlags(bootstrapConfig.Spec.BootstrapCluster.Kubeconfig, false)
	if err != nil {
		return microerror.Mask(err)
	}
	k8sClient.Logger = r.logger

	r.logger.Debugf(ctx, "uploading config")

	content, err := yaml.Marshal(bootstrapConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	err = k8sClient.CreateNamespace(ctx, "giantswarm")
	if err != nil {
		return microerror.Mask(err)
	}

	apps := []client.Object{
		&core.ConfigMap{
			ObjectMeta: meta.ObjectMeta{
				Name:      "capi-bootstrap",
				Namespace: "giantswarm",
			},
			Data: map[string]string{
				"config": string(content),
			},
		},
	}

	err = k8sClient.ApplyResources(ctx, apps)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "uploaded config")

	return nil
}
