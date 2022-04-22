package init

import (
	"io"

	"github.com/giantswarm/micrologger"

	config2 "github.com/giantswarm/capi-bootstrap/pkg/config"
)

type Config struct {
	Logger micrologger.Logger

	Stderr io.Writer
	Stdout io.Writer
}

type Runner struct {
	flag *config2.Flag

	logger micrologger.Logger

	stdout io.Writer
	stderr io.Writer
}
