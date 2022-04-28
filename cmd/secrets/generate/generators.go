package generate

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/capi-bootstrap/pkg/generator"
	"github.com/giantswarm/capi-bootstrap/pkg/generator/config"
	"github.com/giantswarm/capi-bootstrap/pkg/generator/generators/awsiam"
	"github.com/giantswarm/capi-bootstrap/pkg/generator/generators/ca"
	"github.com/giantswarm/capi-bootstrap/pkg/generator/generators/githuboauth"
	"github.com/giantswarm/capi-bootstrap/pkg/generator/generators/lastpass"
	"github.com/giantswarm/capi-bootstrap/pkg/generator/generators/taylorbot"
)

func buildGenerators(config config.Config) (map[string]generator.Generator, error) {
	generators := map[string]generator.Generator{}

	{
		gen, err := awsiam.New(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		generators[awsiam.Name] = gen
	}

	{
		gen, err := ca.New(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		generators[ca.Name] = gen
	}

	{
		gen, err := githuboauth.New(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		generators[githuboauth.Name] = gen
	}

	{
		gen, err := lastpass.New(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		generators[lastpass.Name] = gen
	}

	{
		gen, err := taylorbot.New(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		generators[taylorbot.Name] = gen
	}

	return generators, nil
}
