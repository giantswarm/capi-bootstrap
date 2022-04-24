package fleet

import (
	"github.com/google/go-github/v43/github"

	"github.com/giantswarm/capi-bootstrap/pkg/lastpass"
	"github.com/giantswarm/capi-bootstrap/pkg/repo"
	"github.com/giantswarm/capi-bootstrap/pkg/sops"
)

type Config struct {
	ClusterManifestFile string
	ClusterName         string
}

type Service struct {
	lastpassClient   *lastpass.Client
	repositoryClient *repo.Client
	sopsClient       *sops.Client

	branchName          string
	clusterManifestFile string
	clusterName         string
	defaultBranch       *github.Reference
}
