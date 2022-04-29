package internal

import "github.com/giantswarm/microerror"

var NotFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == NotFoundError
}
