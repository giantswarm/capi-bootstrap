package repo

import (
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v43/github"
)

type Config struct {
	GitAuth      *http.BasicAuth
	GitHubClient *github.Client

	Owner string
	Repo  string
}

type Client struct {
	gitAuth      *http.BasicAuth
	gitHubClient *github.Client

	owner string
	repo  string

	defaultBranch *github.Reference
}

type Change struct {
	Path      string
	Operation string
	Content   string
}

type Commit struct {
	changes []Change
}
