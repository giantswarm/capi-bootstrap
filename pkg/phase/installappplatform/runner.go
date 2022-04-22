package installappplatform

import (
	"context"
	"fmt"
	"path"
	"path/filepath"

	application "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/google/go-github/v43/github"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/capi-bootstrap/pkg/config"
	"github.com/giantswarm/capi-bootstrap/pkg/kubernetes"
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

func (r *Runner) Do(ctx context.Context, environment *config.Environment) error {
	helmClient, err := environment.GetHelmClient()
	if err != nil {
		return microerror.Mask(err)
	}

	opsctlClient, err := environment.GetOpsctlClient()
	if err != nil {
		return microerror.Mask(err)
	}

	k8sClient, err := environment.GetK8sClient()
	if err != nil {
		return microerror.Mask(err)
	}

	gitHubClient, err := environment.GetGitHubClient(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "installing app platform")

	{
		r.logger.Debugf(ctx, "ensuring namespace giantswarm")

		err := k8sClient.CreateNamespace(ctx, "giantswarm")
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "ensured namespace giantswarm")
	}

	// TODO: remove this once draughtsman secret is no longer needed
	{
		r.logger.Debugf(ctx, "ensuring namespace draughtsman")

		err := k8sClient.CreateNamespace(ctx, "draughtsman")
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "ensured namespace draughtsman")

		r.logger.Debugf(ctx, "ensuring draughtsman values configmap and secret")

		resources := []client.Object{
			&core.ConfigMap{
				ObjectMeta: meta.ObjectMeta{
					Name:      "draughtsman-values-configmap",
					Namespace: "draughtsman",
				},
			},
			&core.Secret{
				ObjectMeta: meta.ObjectMeta{
					Name:      "draughtsman-values-secret",
					Namespace: "draughtsman",
				},
			},
		}

		err = k8sClient.ApplyResources(ctx, resources)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "ensured draughtsman values configmap and secret")
	}

	{
		r.logger.Debugf(ctx, "ensuring app platform CRDs")

		crds, err := r.fetchAppPlatformCRDs(ctx, gitHubClient)
		if err != nil {
			return microerror.Mask(err)
		}

		err = k8sClient.ApplyResources(ctx, crds)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "ensured app platform CRDs")
	}

	{
		r.logger.Debugf(ctx, "installing chart-operator")

		err := helmClient.InstallChart("chart-operator", "control-plane-catalog", "giantswarm", "chartOperator.cni.install=true")
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "installed chart-operator")
	}

	{
		r.logger.Debugf(ctx, "installing app-operator")

		err := helmClient.InstallChart("app-operator", "control-plane-catalog", "giantswarm", fmt.Sprintf("provider.kind=%s", environment.ConfigFile.Spec.Provider))
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "installed app-operator")
	}

	{
		r.logger.Debugf(ctx, "ensuring app catalogs")

		err := opsctlClient.EnsureCatalogs()
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "ensured app catalogs")
	}

	{
		r.logger.Debugf(ctx, "ensuring app-operator and chart-operator App CRs")

		apps := []client.Object{
			&core.ConfigMap{
				ObjectMeta: meta.ObjectMeta{
					Name:      "app-operator-user-values",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values": fmt.Sprintf(`provider:
  kind: %s
`, environment.ConfigFile.Spec.Provider),
				},
			},
			&core.ConfigMap{
				ObjectMeta: meta.ObjectMeta{
					Name:      "chart-operator-user-values",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values": `chartOperator:
  cni:
    install: true
`,
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
					Name:      "app-operator",
					Namespace: "giantswarm",
				},
				Spec: application.AppSpec{
					Catalog:          "control-plane-catalog",
					CatalogNamespace: "giantswarm",
					KubeConfig: application.AppSpecKubeConfig{
						InCluster: true,
					},
					Name:      "app-operator",
					Namespace: "giantswarm",
					UserConfig: application.AppSpecUserConfig{
						ConfigMap: application.AppSpecUserConfigConfigMap{
							Name:      "app-operator-user-values",
							Namespace: "giantswarm",
						},
					},
					Version: "5.8.0",
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
					Name:      "chart-operator",
					Namespace: "giantswarm",
				},
				Spec: application.AppSpec{
					Catalog:          "control-plane-catalog",
					CatalogNamespace: "giantswarm",
					KubeConfig: application.AppSpecKubeConfig{
						InCluster: true,
					},
					Name:      "chart-operator",
					Namespace: "giantswarm",
					UserConfig: application.AppSpecUserConfig{
						ConfigMap: application.AppSpecUserConfigConfigMap{
							Name:      "chart-operator-user-values",
							Namespace: "giantswarm",
						},
					},
					Version: "2.20.1",
				},
			},
		}

		err := k8sClient.ApplyResources(ctx, apps)
		if err != nil {
			return microerror.Mask(err)
		}

		err = k8sClient.WaitForAppsDeployed(ctx, "giantswarm", []string{
			"app-operator",
			"chart-operator",
		})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "ensured app-operator and chart-operator App CRs")
	}

	return nil
}

func (r *Runner) fetchAppPlatformCRDs(ctx context.Context, gitHubClient *github.Client) ([]client.Object, error) {
	owner := "giantswarm"
	repo := "apiextensions-application"

	latestRelease, _, err := gitHubClient.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	getOptions := github.RepositoryContentGetOptions{
		Ref: "refs/tags/" + *latestRelease.TagName,
	}

	crdPath := path.Join("config", "crd")
	_, contents, _, err := gitHubClient.Repositories.GetContents(ctx, "giantswarm", "apiextensions-application", crdPath, &getOptions)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var crds []client.Object
	for _, file := range contents {
		if filepath.Ext(*file.Name) != ".yaml" {
			continue
		}

		filePath := path.Join(crdPath, *file.Name)
		contentReader, _, err := gitHubClient.Repositories.DownloadContents(ctx, "giantswarm", "apiextensions-application", filePath, &getOptions)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		entryCRDs, err := kubernetes.DecodeCRDs(contentReader)
		if err != nil {
			return nil, err
		}

		for _, crd := range entryCRDs {
			crds = append(crds, crd)
		}
	}

	return crds, nil
}
