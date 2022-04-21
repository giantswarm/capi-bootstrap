package launch

import (
	"context"
	"io"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	config2 "github.com/giantswarm/capi-bootstrap/pkg/config"
	"github.com/giantswarm/capi-bootstrap/pkg/helm"
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

	typedClient, err := kubernetes.TypedClientFromFlags(bootstrapConfig.Spec.BootstrapCluster.Kubeconfig, false)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "launching capi-bootstrap job")

	helmClient := helm.Client{
		KubeconfigPath: bootstrapConfig.Spec.BootstrapCluster.Kubeconfig,
	}
	err = helmClient.InstallChart("capi-bootstrap", "control-plane-catalog", "giantswarm", "")
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "launched capi-bootstrap job")

	var podList core.PodList
	err = k8sClient.Client.List(ctx, &podList, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": "capi-bootstrap",
		}),
		Namespace: "giantswarm",
		Limit:     1,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "found capi-bootstrap pod %s", podList.Items[0].Name)

	request := typedClient.CoreV1().Pods("giantswarm").GetLogs(podList.Items[0].Name, &core.PodLogOptions{
		Container:  "capi-bootstrap",
		Follow:     true,
		LimitBytes: nil,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	podLogs, err := request.Stream(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	defer podLogs.Close()

	r.logger.Debugf(ctx, "following pod logs")

	for {
		buffer := make([]byte, 100)
		bytesRead, err := podLogs.Read(buffer)
		if bytesRead > 0 {
			r.logger.Debugf(ctx, string(buffer[:bytesRead]))
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
