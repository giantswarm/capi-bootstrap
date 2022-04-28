package generate

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagBaseDomain  = "base-domain"
	flagClusterName = "cluster-name"
	flagProvider    = "provider"
)

type flags struct {
	BaseDomain  string
	ClusterName string
	Provider    string
}

func (f *flags) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.BaseDomain, flagBaseDomain, "", `Base domain (usually <customer>.gigantic.io)`)
	cmd.Flags().StringVar(&f.ClusterName, flagClusterName, "", `Management cluster name`)
	cmd.Flags().StringVar(&f.Provider, flagProvider, "", `Infrastructure provider for management cluster`)
}

func (f *flags) Validate() error {
	if f.BaseDomain == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagBaseDomain)
	}
	if f.ClusterName == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagClusterName)
	}
	if f.Provider == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagProvider)
	}

	return nil
}
