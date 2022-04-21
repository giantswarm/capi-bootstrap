package pivot

import (
	"io"

	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/capi-bootstrap/pkg/kubernetes"
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

type ClusterScope struct {
	K8sClient      *kubernetes.Client
	KubeconfigPath string
}
