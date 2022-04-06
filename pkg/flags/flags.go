package flags

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/giantswarm/capi-bootstrap/pkg/shell"
	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

type AppReference struct {
	appName string
	catalog string
	version string
}

type Flags struct {
	Command string

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

	Auth0ctlVersion   string
	ClusterctlVersion string
	DevctlVersion     string

	ClusterApp       AppReference
	DefaultAppsApp   AppReference
	ClusterResources AppReference
	CNI              AppReference
	CloudProvider    AppReference
	CertManager      AppReference
	Kyverno          AppReference

	Input string
}

var providers = []string{
	"aws",
	"azure",
	"gcp",
	"openstack",
	"vsphere",
}

func (f *Flags) Validate() error {
	if f.ManagementClusterName == "" {
		log.Fatal("--cluster-name is required")
	}
	if f.CustomerBaseDomain == "" {
		log.Fatal("--customer-base-domain is required")
	}
	if f.Provider == "" {
		log.Fatal("--provider is required")
	}
	if !util.Contains(providers, f.Provider) {
		log.Fatal("--provider should be one of: ", strings.Join(providers, ", "))
	}
	if f.TeamName == "" {
		log.Fatal("--team-name is required")
	}

	if os.Getenv("GITHUB_TOKEN") == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	err := shell.VerifyBinaryExists("helm")
	if err != nil {
		return err
	}

	err = shell.VerifyBinaryExists("lpass")
	if err != nil {
		return err
	}

	err = shell.VerifyBinaryExists("kind")
	if err != nil {
		return err
	}

	err = shell.VerifyBinaryExists("opsctl")
	if err != nil {
		return err
	}

	if f.Input == "" {
		log.Fatal("--file is required")
	} else if f.Input == "-" {
		info, err := os.Stdin.Stat()
		if err != nil {
			return err
		}

		if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
			return errors.New("usage: kubectl gs template cluster | capi-bootstrap")
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

		f.Input = string(output)
	} else {
		input, err := os.ReadFile(f.Input)
		if err != nil {
			return err
		}

		f.Input = string(input)
	}

	return nil
}
