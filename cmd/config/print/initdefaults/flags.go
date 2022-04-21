package initdefaults

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/capi-bootstrap/pkg/key"
	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

const (
	flagLastPassUsername   = "lastpass-username"
	flagLastPassPassword   = "lastpass-password"
	flagLastPassTOTPSecret = "lastpass-totp-secret"

	flagCustomerBaseDomain    = "customer-base-domain"
	flagCustomer              = "customer"
	flagManagementClusterName = "management-cluster-name"
	flagClusterNamespace      = "cluster-namespace"
	flagKindClusterName       = "kind-cluster-name"
	flagProvider              = "provider"
	flagTeamName              = "team-name"

	flagProduction = "production"

	flagFile = "file"
)

type Flags struct {
	LastPassUsername   string
	LastPassPassword   string
	LastPassTOTPSecret string

	CustomerBaseDomain    string
	Customer              string
	ManagementClusterName string
	ClusterNamespace      string
	KindClusterName       string
	Provider              string
	TeamName              string

	Production bool

	File string
}

func (f *Flags) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.LastPassPassword, flagLastPassPassword, "", `Existing release upon which to base the new release. Must follow semver format.`)
	cmd.Flags().StringVar(&f.LastPassUsername, flagLastPassUsername, "", `Existing release upon which to base the new release. Must follow semver format.`)
	cmd.Flags().StringVar(&f.LastPassTOTPSecret, flagLastPassTOTPSecret, "", `Existing release upon which to base the new release. Must follow semver format.`)

	cmd.Flags().StringVar(&f.Customer, flagCustomer, "", `Existing release upon which to base the new release. Must follow semver format.`)
	cmd.Flags().StringVar(&f.CustomerBaseDomain, flagCustomerBaseDomain, "", `Existing release upon which to base the new release. Must follow semver format.`)
	cmd.Flags().StringVar(&f.ManagementClusterName, flagManagementClusterName, "", `Existing release upon which to base the new release. Must follow semver format.`)
	cmd.Flags().StringVar(&f.ClusterNamespace, flagClusterNamespace, "", `Existing release upon which to base the new release. Must follow semver format.`)
	cmd.Flags().StringVar(&f.KindClusterName, flagKindClusterName, "", `Existing release upon which to base the new release. Must follow semver format.`)
	cmd.Flags().StringVar(&f.Provider, flagProvider, "", `Existing release upon which to base the new release. Must follow semver format.`)
	cmd.Flags().StringVar(&f.TeamName, flagTeamName, "", `Existing release upon which to base the new release. Must follow semver format.`)

	cmd.Flags().BoolVar(&f.Production, flagProduction, false, `Existing release upon which to base the new release. Must follow semver format.`)

	cmd.Flags().StringVarP(&f.File, flagFile, "f", "", `Existing release upon which to base the new release. Must follow semver format.`)
}

func (f *Flags) Validate() error {
	if f.LastPassPassword == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagLastPassPassword)
	}
	if f.LastPassUsername == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagLastPassUsername)
	}
	if f.CustomerBaseDomain == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagCustomerBaseDomain)
	}
	if f.Customer == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagCustomer)
	}
	if f.ManagementClusterName == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagManagementClusterName)
	}
	if f.ClusterNamespace == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagClusterNamespace)
	}
	if f.KindClusterName == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagKindClusterName)
	}
	if f.Customer == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagCustomer)
	}
	if f.Provider == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagProvider)
	}
	if !util.Contains(key.AllowedProviders, f.Provider) {
		log.Fatal("--provider should be one of: ", strings.Join(key.AllowedProviders, ", "))
	}
	if f.TeamName == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagTeamName)
	}
	if f.File == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagFile)
	}

	if f.File == "" {
		log.Fatal("--file is required")
	} else if f.File == "-" {
		info, err := os.Stdin.Stat()
		if err != nil {
			return microerror.Mask(err)
		}

		if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
			return microerror.Mask(errors.New("usage: kubectl gs template cluster | capi-bootstrap"))
		}

		reader := bufio.NewReader(os.Stdin)
		var output []rune

		for {
			input, _, err := reader.ReadRune()
			if err != nil && err == io.EOF {
				break
			}
			output = append(output, input)
		}

		f.File = string(output)
	} else {
		input, err := os.ReadFile(f.File)
		if err != nil {
			return microerror.Mask(err)
		}

		f.File = string(input)
	}

	return nil
}
