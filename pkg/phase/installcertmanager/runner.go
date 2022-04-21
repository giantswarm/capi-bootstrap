package installcertmanager

import (
	"context"

	application "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/capi-bootstrap/pkg/kubernetes"
)

func (r *Runner) Run(cmd *cobra.Command, _ []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context())
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Runner) Do(ctx context.Context) error {
	k8sClient, err := kubernetes.ClientFromFlags(r.flag.Kubeconfig, r.flag.InCluster)
	if err != nil {
		return microerror.Mask(err)
	}
	k8sClient.Logger = r.logger

	r.logger.Debugf(ctx, "installing cert-manager")

	apps := []client.Object{
		&application.App{
			ObjectMeta: meta.ObjectMeta{
				Annotations: map[string]string{
					"chart-operator.giantswarm.io/force-helm-upgrade": "true",
				},
				Labels: map[string]string{
					"app-operator.giantswarm.io/version": "0.0.0",
				},
				Name:      "cert-manager",
				Namespace: "giantswarm",
			},
			Spec: application.AppSpec{
				Catalog:          "control-plane-catalog",
				CatalogNamespace: "giantswarm",
				KubeConfig: application.AppSpecKubeConfig{
					InCluster: true,
				},
				Name:      "cert-manager-app",
				Namespace: "giantswarm",
				Version:   "2.12.0",
			},
		},
	}

	err = k8sClient.ApplyResources(ctx, apps)
	if err != nil {
		return microerror.Mask(err)
	}

	err = k8sClient.WaitForAppsDeployed(ctx, "giantswarm", []string{
		"cert-manager",
	})
	if err != nil {
		return microerror.Mask(err)
	}

	var deploymentKeys []client.ObjectKey
	for _, name := range []string{
		"cert-manager-controller",
		"cert-manager-webhook",
		"cert-manager-cainjector",
	} {
		deploymentKeys = append(deploymentKeys, client.ObjectKey{
			Namespace: "giantswarm",
			Name:      name,
		})
	}
	err = k8sClient.WaitForDeployments(ctx, deploymentKeys)
	if err != nil {
		return microerror.Mask(err)
	}

	err = k8sClient.WaitForCRDs(ctx, []string{
		"certificaterequests.cert-manager.io",
		"certificates.cert-manager.io",
		"clusterissuers.cert-manager.io",
		"issuers.cert-manager.io",
	})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "installed cert-manager")

	return nil
}
