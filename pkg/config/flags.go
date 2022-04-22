package config

import (
	"io/fs"
	"os"
	"strings"

	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

const (
	flagConfigFile      = "config-file"
	flagSecretDirectory = "secret-directory"

	flagEnvironment = "environment"
	flagInCluster   = "in-cluster"
)

type Flag struct {
	ConfigFile      string
	SecretDirectory string

	Environment string
	InCluster   bool
}

func (f *Flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.ConfigFile, flagConfigFile, "", "")
	cmd.Flags().StringVar(&f.SecretDirectory, flagSecretDirectory, "", "")

	cmd.Flags().StringVar(&f.Environment, flagEnvironment, "", "")
	cmd.Flags().BoolVar(&f.InCluster, flagInCluster, false, "")
}

var environments = []string{
	"bootstrap",
	"permanent",
}

func (f *Flag) Validate() error {
	if f.ConfigFile == "" {
		return microerror.Maskf(invalidFlagError, "--%s is required", flagConfigFile)
	}
	if f.InCluster && f.SecretDirectory == "" {
		return microerror.Maskf(invalidFlagError, "--%s is required", flagSecretDirectory)
	}
	if f.InCluster && f.Environment == "" {
		return microerror.Maskf(invalidFlagError, "--%s is required", flagEnvironment)
	}
	if f.InCluster && !util.Contains(environments, f.Environment) {
		return microerror.Maskf(invalidFlagError, "--%s must be one of: ", strings.Join(environments, ", "))
	}

	return nil
}

func (f *Flag) BuildEnvironment(logger micrologger.Logger) (*Environment, error) {
	content, err := os.ReadFile(f.ConfigFile)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var configFile ConfigFile
	err = yaml.Unmarshal(content, &configFile)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = configFile.Validate()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var secrets map[string]string
	if f.InCluster {
		secrets, err = loadDirectoryFiles(os.DirFS(f.SecretDirectory))
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return &Environment{
		Logger:     logger,
		InCluster:  f.InCluster,
		Type:       f.Environment,
		ConfigFile: configFile,
		Secrets:    secrets,
	}, nil
}

func loadDirectoryFiles(directory fs.FS) (map[string]string, error) {
	result := map[string]string{}

	err := fs.WalkDir(directory, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		}

		content, err := fs.ReadFile(directory, path)
		if err != nil {
			return microerror.Mask(err)
		}

		result[path] = string(content)

		return nil
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return result, nil
}
