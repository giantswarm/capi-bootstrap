package initdefaults

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

func (r *Runner) Run(cmd *cobra.Command, args []string) error {
	err := r.flags.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context(), cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Runner) Do(_ context.Context, _ *cobra.Command, _ []string) error {
	config := BootstrapConfig{
		APIVersion: "v1alpha1",
		Kind:       "BootstrapConfig",
		Spec: BootstrapConfigSpec{
			BootstrapCluster: BootstrapCluster{
				Name: r.flags.ManagementClusterName,
			},
			PermanentCluster: PermanentCluster{
				Name: r.flags.ManagementClusterName,
			},
			Config: BootstrapConfigSpecConfig{
				AppCollection: AppCollection{
					BranchName: fmt.Sprintf("%s_auto_config", r.flags.ManagementClusterName),
				},
				Installations: Installations{
					BranchName: fmt.Sprintf("%s_auto_config", r.flags.ManagementClusterName),
				},
			},
			Kubeconfig: Kubeconfig{
				Group: fmt.Sprintf("Shared-%s/%s", r.flags.TeamName, r.flags.Customer),
				Name:  fmt.Sprintf("%s.kubeconfig", r.flags.ManagementClusterName),
			},
			Provider: r.flags.Provider,
		},
	}
	if r.flags.Provider == "openstack" {
		config.Spec.LastpassSecrets = append(config.Spec.LastpassSecrets, LastpassSecret{
			Group:           fmt.Sprintf("Shared-%s/%s", r.flags.TeamName, r.flags.Customer),
			Name:            "default-giantswarm-openrc.sh",
			SecretName:      "cloud-config",
			SecretNamespace: r.flags.ClusterNamespace,
		})
	}
	output, err := yaml.Marshal(config)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = r.stdout.Write(output)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
