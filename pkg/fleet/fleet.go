package fleet

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/lastpass"
	"github.com/giantswarm/capi-bootstrap/pkg/repo"
	"github.com/giantswarm/capi-bootstrap/pkg/sops"
	"github.com/giantswarm/capi-bootstrap/pkg/util"
)

const (
	owner     = "giantswarm"
	fleetRepo = "management-clusters-fleet-openstack"
)

func mustLookupEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", microerror.Maskf(invalidConfigError, "%s must be defined", key)
	}
	return value, nil
}

func New(config Config) (*Service, error) {
	gitHubToken, err := mustLookupEnv("GITHUB_TOKEN")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var gitHubClient *github.Client
	{
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: gitHubToken,
		})
		httpClient := oauth2.NewClient(context.Background(), tokenSource)
		gitHubClient = github.NewClient(httpClient)
	}

	var sopsClient *sops.Client
	{
		username, err := mustLookupEnv("LASTPASS_USERNAME")
		if err != nil {
			return nil, microerror.Mask(err)
		}
		password, err := mustLookupEnv("LASTPASS_PASSWORD")
		if err != nil {
			return nil, microerror.Mask(err)
		}
		totpSecret, err := mustLookupEnv("LASTPASS_TOTP_SECRET")
		if err != nil {
			return nil, microerror.Mask(err)
		}
		lastpassClient, err := lastpass.New(lastpass.Config{
			Username:   username,
			Password:   password,
			TOTPSecret: totpSecret,
		})
		if err != nil {
			return nil, microerror.Mask(err)
		}

		sopsClient, err = sops.New(sops.Config{
			LastpassClient: lastpassClient,
			ClusterName:    config.ClusterName,
		})
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	gitAuth := http.BasicAuth{
		Username: gitHubToken,
		Password: "",
	}

	repository, err := repo.New(repo.Config{
		GitAuth:      &gitAuth,
		GitHubClient: gitHubClient,
		Owner:        owner,
		Repo:         fleetRepo,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &Service{
		sopsClient:       sopsClient,
		repositoryClient: repository,

		branchName:          fmt.Sprintf("bootstrap-%s", config.ClusterName),
		clusterManifestFile: config.ClusterManifestFile,
		clusterName:         config.ClusterName,
	}, nil
}

func (s *Service) EnsureCreated(ctx context.Context) error {
	err := s.sopsClient.EnsureEncryptionKey(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	err = s.ensureFleetCreated(ctx)
	return microerror.Mask(err)
}

func (s *Service) EnsureDeleted(ctx context.Context) error {
	err := s.ensureFleetDeleted(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	err = s.sopsClient.DeleteEncryptionKey(ctx)
	return microerror.Mask(err)
}

func (s *Service) ensureEncryptionKey(worktree billy.Filesystem) (*github.TreeEntry, error) {
	filePath := fmt.Sprintf("clusters/%s/.sops.yaml", s.clusterName)
	content, err := s.sopsClient.RenderConfig()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	entry, err := s.ensureFile(worktree, filePath, content)
	return entry, microerror.Mask(err)
}

func secretsMatch(left *core.Secret, right *core.Secret) (bool, error) {
	leftYAML, err := yaml.Marshal(left)
	if err != nil {
		return false, microerror.Mask(err)
	}
	rightYAML, err := yaml.Marshal(right)
	if err != nil {
		return false, microerror.Mask(err)
	}
	return string(leftYAML) == string(rightYAML), nil
}

func (s *Service) ensureSecret(worktree billy.Filesystem, filePath string, expectedSecret *core.Secret) (*github.TreeEntry, error) {
	if _, err := worktree.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		// fall through
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else {
		file, err := worktree.Open(filePath)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		encrypted, err := io.ReadAll(file)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		decryptedSecret, err := s.sopsClient.DecryptSecret(encrypted)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if match, err := secretsMatch(expectedSecret, decryptedSecret); err != nil {
			return nil, microerror.Mask(err)
		} else if match {
			return nil, nil
		}
	}

	encrypted, err := s.sopsClient.EncryptSecret(expectedSecret)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &github.TreeEntry{
		Path:    github.String(filePath),
		Mode:    github.String("100644"),
		Type:    github.String("blob"),
		Content: github.String(string(encrypted)),
	}, nil
}

func (s *Service) ensureFile(worktree billy.Filesystem, filePath, expectedContent string) (*github.TreeEntry, error) {
	if _, err := worktree.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		// fall through
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else {
		file, err := worktree.Open(filePath)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		content, err := io.ReadAll(file)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if string(content) == expectedContent {
			return nil, nil
		}
	}

	return &github.TreeEntry{
		Path:    github.String(filePath),
		Mode:    github.String("100644"),
		Type:    github.String("blob"),
		Content: github.String(expectedContent),
	}, nil
}

func (s *Service) ensureCloudConfigSecret(ctx context.Context, worktree billy.Filesystem) (*github.TreeEntry, error) {
	filePath := fmt.Sprintf("clusters/%s/cloud-config-secret.yaml", s.clusterName)
	openrcSecret, err := s.lastpassClient.GetAccount(ctx, "Shared-Team Rocket", "CAPO\\THG", "default-openrc.sh")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cloudConfig := util.OpenrcToCloudConfig(openrcSecret.Notes)
	cloudConfigYAML, err := yaml.Marshal(cloudConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cloudConfigEncoded := base64.StdEncoding.EncodeToString(cloudConfigYAML)

	secret := core.Secret{
		TypeMeta: meta.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      "cloud-config",
			Namespace: "giantswarm",
		},
		Data: map[string][]byte{
			"clouds.yaml": []byte(cloudConfigEncoded),
		},
	}

	entry, err := s.ensureSecret(worktree, filePath, &secret)
	return entry, microerror.Mask(err)
}

func (s *Service) buildFleetCreateEntries(ctx context.Context, worktree billy.Filesystem) ([]*github.TreeEntry, error) {
	var entries []*github.TreeEntry

	if entry, err := s.ensureEncryptionKey(worktree); err != nil {
		return nil, microerror.Mask(err)
	} else if entry != nil {
		entries = append(entries, entry)
	}

	if entry, err := s.ensureCloudConfigSecret(ctx, worktree); err != nil {
		return nil, microerror.Mask(err)
	} else if entry != nil {
		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *Service) ensureFleetCreated(ctx context.Context) error {
	var baseRef *github.Reference
	if branchRef, err := s.repositoryClient.GetBranchReference(ctx, s.branchName); IsNotFound(err) {
		baseRef, err = s.repositoryClient.GetDefaultBranchReference(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		baseRef = branchRef
	}

	worktree, err := s.repositoryClient.GetWorktree(ctx, baseRef.GetRef())
	if err != nil {
		return microerror.Mask(err)
	}

	entries, err := s.buildFleetCreateEntries(ctx, worktree)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(entries) == 0 {
		// no changes necessary
		return nil
	}

	var commit repo.Commit
	_, err = s.repositoryClient.PushCommit(ctx, s.branchName, commit)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = s.repositoryClient.CreateOrUpdatePullRequest(ctx, github.NewPullRequest{
		Body:                github.String("PR generated by capi-bootstrap"),
		Draft:               github.Bool(false),
		MaintainerCanModify: github.Bool(true),
		Title:               github.String("Bootstrap " + s.clusterName),
	})
	return microerror.Mask(err)
}

func (s *Service) ensureFleetDeleted(ctx context.Context) error {
	if existing, err := s.repositoryClient.GetPullRequest(ctx, s.branchName); repo.IsNotFound(err) {
		err = s.repositoryClient.ClosePullRequest(ctx, existing.GetNumber())
		if err != nil {
			return microerror.Mask(err)
		}
	} else if err != nil {
		return microerror.Mask(err)
	}

	defaultBranchRef, err := s.repositoryClient.GetDefaultBranchReference(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	worktree, err := s.repositoryClient.GetWorktree(ctx, defaultBranchRef.GetRef())
	if err != nil {
		return microerror.Mask(err)
	}

	filesToDelete, err := worktree.ReadDir("clusters/%s")
	if err != nil {
		return microerror.Mask(err)
	}

	if len(filesToDelete) == 0 {
		return nil
	}

	var entries []*github.TreeEntry
	for _, file := range filesToDelete {
		entries = append(entries, &github.TreeEntry{
			Path: github.String(file.Name()),
			Mode: github.String(file.Mode().String()),
			Type: github.String(""),
		})
	}

	var commit repo.Commit
	_, err = s.repositoryClient.PushCommit(ctx, s.branchName, commit)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = s.repositoryClient.CreateOrUpdatePullRequest(ctx, github.NewPullRequest{
		Body:                github.String("PR generated by capi-bootstrap"),
		Draft:               github.Bool(false),
		MaintainerCanModify: github.Bool(true),
		Title:               github.String("Bootstrap " + s.clusterName),
	})
	return microerror.Mask(err)
}
