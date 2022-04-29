package lastpass

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/capi-bootstrap/pkg/lastpass/internal"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == internal.NotFoundError
}

var commandFailedError = &microerror.Error{
	Kind: "commandFailedError",
}

// IsCommandFailed asserts commandFailedError.
func IsCommandFailed(err error) bool {
	return microerror.Cause(err) == commandFailedError
}
