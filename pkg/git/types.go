package git

import "github.com/google/go-github/v43/github"

type Config struct {
	GitHubClient *github.Client

	Owner string
	Repo  string

	BranchName     string
	CommitAuthor   github.CommitAuthor
	CommitTemplate github.Commit
}

type Client struct {
	gitHubClient *github.Client

	owner string
	repo  string

	branchName     string
	commitAuthor   github.CommitAuthor
	commitTemplate github.Commit

	defaultBranch *github.Reference
}
