package installclusterappsoperator

import (
	"context"
	"fmt"

	application "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/capi-bootstrap/pkg/kubernetes"
)

func (r *Runner) Run(cmd *cobra.Command, args []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context())
	return microerror.Mask(err)
}

func (r *Runner) Do(ctx context.Context) error {
	k8sClient, err := kubernetes.ClientFromFlags(r.flag.Kubeconfig, r.flag.InCluster)
	if err != nil {
		return microerror.Mask(err)
	}
	k8sClient.Logger = r.logger

	{
		r.logger.Debugf(ctx, "installing cluster-apps-operator")

		apps := []client.Object{
			&core.ConfigMap{
				ObjectMeta: meta.ObjectMeta{
					Name:      "cluster-apps-operator-user-values",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values": fmt.Sprintf(`baseDomain: %s
provider:
  kind: %s
`, r.flag.BaseDomain, r.flag.Provider),
				},
			},
			&application.App{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"chart-operator.giantswarm.io/force-helm-upgrade": "true",
					},
					Labels: map[string]string{
						"app-operator.giantswarm.io/version": "0.0.0",
					},
					Name:      "cluster-apps-operator",
					Namespace: "giantswarm",
				},
				Spec: application.AppSpec{
					Catalog:          "control-plane-catalog",
					CatalogNamespace: "giantswarm",
					KubeConfig: application.AppSpecKubeConfig{
						InCluster: true,
					},
					Name:      "cluster-apps-operator",
					Namespace: "giantswarm",
					UserConfig: application.AppSpecUserConfig{
						ConfigMap: application.AppSpecUserConfigConfigMap{
							Name:      "cluster-apps-operator-user-values",
							Namespace: "giantswarm",
						},
					},
					Version: "1.5.0",
				},
			},
		}

		err := k8sClient.ApplyResources(ctx, apps)
		if err != nil {
			return microerror.Mask(err)
		}

		err = k8sClient.WaitForAppsDeployed(ctx, "giantswarm", []string{
			"cluster-apps-operator",
		})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "ensured cluster-apps-operator")
	}

	return nil
}
