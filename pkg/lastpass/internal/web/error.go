package web

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var unauthenticatedError = &microerror.Error{
	Kind: "unauthenticatedError",
}

// IsUnauthenticated asserts unauthenticatedError.
func IsUnauthenticated(err error) bool {
	return microerror.Cause(err) == unauthenticatedError
}
