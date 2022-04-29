package sops

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var commandFailedError = &microerror.Error{
	Kind: "commandFailedError",
}

// IsCommandFailed asserts commandFailedError.
func IsCommandFailed(err error) bool {
	return microerror.Cause(err) == commandFailedError
}
