package encrypt

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
	flag *flags

	logger micrologger.Logger

	stdout io.Writer
	stderr io.Writer
}
