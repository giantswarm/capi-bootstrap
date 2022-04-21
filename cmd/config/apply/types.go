package apply

import (
	"io"

	"github.com/giantswarm/micrologger"
)

type Runner struct {
	flags *flags

	logger micrologger.Logger

	stdout io.Writer
	stderr io.Writer
}

type Config struct {
	Logger micrologger.Logger

	Stderr io.Writer
	Stdout io.Writer
}
