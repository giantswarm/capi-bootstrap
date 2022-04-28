package generate

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagBaseDomain  = "base-domain"
	flagClusterName = "cluster-name"
	flagCustomer    = "customer"
	flagPipeline    = "pipeline"

	flagInstallationSecretsFile = "installation-secrets-file"
	flagProvider                = "provider"
	flagOutputDirectory         = "output-directory"
)

type flags struct {
	BaseDomain  string
	ClusterName string
	Customer    string
	Pipeline    string
	Provider    string

	InstallationSecretsFile string
	OutputDirectory         string
}

func (f *flags) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.BaseDomain, flagBaseDomain, "", `Management cluster infrastructure provider`)
	cmd.Flags().StringVar(&f.ClusterName, flagClusterName, "", `Management cluster infrastructure provider`)
	cmd.Flags().StringVar(&f.Customer, flagCustomer, "", `Management cluster infrastructure provider`)
	cmd.Flags().StringVar(&f.Pipeline, flagPipeline, "", `Management cluster infrastructure provider`)
	cmd.Flags().StringVar(&f.Provider, flagProvider, "", `Management cluster infrastructure provider`)

	cmd.Flags().StringVar(&f.InstallationSecretsFile, flagInstallationSecretsFile, "", `Path to file containing installation secrets`)
	cmd.Flags().StringVar(&f.OutputDirectory, flagOutputDirectory, "", `Directory in which to write the generated config files`)
}

func (f *flags) Validate() error {
	if f.InstallationSecretsFile == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagInstallationSecretsFile)
	}
	if f.Provider == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagProvider)
	}
	if f.OutputDirectory == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagOutputDirectory)
	}

	return nil
}
