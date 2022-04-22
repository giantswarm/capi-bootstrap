package kubernetes

import (
	"context"

	"github.com/giantswarm/microerror"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

func (c Client) CreateCloudConfigSecret(ctx context.Context, openrcContent string, clusterNamespace string) error {
	cloudConfig := util.OpenrcToCloudConfig(openrcContent)
	cloudConfigYAML, err := yaml.Marshal(cloudConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	secret := core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Labels: map[string]string{
				"clusterctl.cluster.x-k8s.io/move": "true",
			},
			Name:      "cloud-config",
			Namespace: clusterNamespace,
		},
		StringData: map[string]string{
			"clouds.yaml": string(cloudConfigYAML),
		},
	}

	err = c.ApplyResources(ctx, []client.Object{&secret})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
