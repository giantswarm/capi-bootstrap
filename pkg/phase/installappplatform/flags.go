package installappplatform

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

const (
	flagConfigFile = "config-file"
	flagInCluster  = "in-cluster"
	flagTarget     = "target"
)

type flags struct {
	ConfigFile string
	InCluster  bool
	Target     string
}

func (f *flags) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.ConfigFile, flagConfigFile, "", "")
	cmd.Flags().BoolVar(&f.InCluster, flagInCluster, false, "")
	cmd.Flags().StringVar(&f.Target, flagTarget, "", "")
}

func (f *flags) Validate() error {
	if f.ConfigFile == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagConfigFile)
	}
	if f.Target == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagTarget)
	}
	targets := []string{"bootstrap", "permanent"}
	if util.Contains(targets, f.Target) {
		return microerror.Maskf(invalidFlagError, "--%s must be one of: ", strings.Join(targets, ", "))
	}

	return nil
}
