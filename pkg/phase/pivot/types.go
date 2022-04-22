package pivot

import (
	"io"

	"github.com/giantswarm/micrologger"

	config2 "github.com/giantswarm/capi-bootstrap/pkg/config"
	"github.com/giantswarm/capi-bootstrap/pkg/kubernetes"
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

type ClusterScope struct {
	K8sClient      *kubernetes.Client
	KubeconfigPath string
}
