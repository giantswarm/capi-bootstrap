package writesopsconfig

import (
	"io"

	"github.com/giantswarm/micrologger"
)

type Config struct {
	Logger micrologger.Logger

	Stderr io.Writer
	Stdout io.Writer
}

type Runner struct {
	logger micrologger.Logger

	stdout io.Writer
	stderr io.Writer
}
