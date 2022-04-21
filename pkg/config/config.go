package config

import (
	"os"
	"strings"

	"github.com/giantswarm/microerror"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/key"
	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

func FromFile(file string) (BootstrapConfig, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return BootstrapConfig{}, microerror.Mask(err)
	}

	var bootstrapConfig BootstrapConfig
	err = yaml.Unmarshal(content, &bootstrapConfig)
	if err != nil {
		return BootstrapConfig{}, microerror.Mask(err)
	}

	return bootstrapConfig, nil
}

func (b BootstrapConfig) ToFile(file string) error {
	content, err := yaml.Marshal(b)
	if err != nil {
		return microerror.Mask(err)
	}

	err = os.WriteFile(file, content, 0644)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (b BootstrapConfig) Validate() error {
	if b.Kind != "BootstrapConfig" {
		return microerror.Maskf(invalidFlagError, "invalid kind %s", b.Kind)
	}
	if b.APIVersion != "v1alpha1" {
		return microerror.Maskf(invalidFlagError, "invalid api version %s", b.APIVersion)
	}
	if !util.Contains(key.AllowedProviders, b.Spec.Provider) {
		return microerror.Maskf(invalidFlagError, "provider must be one of: %s", strings.Join(key.AllowedProviders, ","))
	}
	return nil
}
