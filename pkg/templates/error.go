package templates

import "github.com/giantswarm/microerror"

var invalidTemplateError = &microerror.Error{
	Kind: "invalidTemplateError",
}

// IsInvalidTemplate asserts invalidTemplateError.
func IsInvalidTemplate(err error) bool {
	return microerror.Cause(err) == invalidTemplateError
}
