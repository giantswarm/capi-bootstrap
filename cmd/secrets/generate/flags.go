package generate

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	flagBaseDomain  = "base-domain"
	flagClusterName = "cluster-name"
	flagEncrypt     = "encrypt"
	flagProvider    = "provider"
)

type flags struct {
	Encrypt bool

	BaseDomain  string
	ClusterName string
	Provider    string
}

func (f *flags) Init(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&f.Encrypt, flagEncrypt, true, `If true, output will be encrypted using sops`)

	cmd.Flags().StringVar(&f.BaseDomain, flagBaseDomain, "", `Base domain (usually <customer>.gigantic.io or test.gigantic.io for test installations)`)
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
