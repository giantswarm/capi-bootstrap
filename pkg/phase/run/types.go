package run

import (
	"io"

	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/capi-bootstrap/pkg/config"
)

type Config struct {
	Logger micrologger.Logger

	Stderr io.Writer
	Stdout io.Writer
}

type Runner struct {
	flag *config.Flag

	logger micrologger.Logger

	stdout io.Writer
	stderr io.Writer
}
