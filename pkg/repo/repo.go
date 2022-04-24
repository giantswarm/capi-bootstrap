package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-github/v43/github"
)

func New(config Config) (*Client, error) {
	return &Client{
		gitAuth:      config.GitAuth,
		gitHubClient: config.GitHubClient,

		owner: config.Owner,
		repo:  config.Repo,
	}, nil
}

func (s *Client) GetDefaultBranchReference(ctx context.Context) (*github.Reference, error) {
	if s.defaultBranch == nil {
		repo, _, err := s.gitHubClient.Repositories.Get(ctx, s.owner, s.repo)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		refName := branchNameToHeadRef(repo.GetDefaultBranch())
		s.defaultBranch, _, err = s.gitHubClient.Git.GetRef(ctx, s.owner, s.repo, refName)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return s.defaultBranch, nil
}

func (s *Client) HasBranch(ctx context.Context, branchName string) (bool, error) {
	matches, _, err := s.gitHubClient.Git.ListMatchingRefs(ctx, s.owner, s.repo, &github.ReferenceListOptions{
		Ref: branchNameToHeadRef(branchName),
	})
	return len(matches) > 0, microerror.Mask(err)
}

func (s *Client) GetBranchReference(ctx context.Context, branchName string) (*github.Reference, error) {
	branchRef, _, err := s.gitHubClient.Git.GetRef(ctx, s.owner, s.repo, branchNameToHeadRef(branchName))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return branchRef, nil
}

func (s *Client) GetOrCreateBranch(ctx context.Context, branchName string) (*github.Reference, error) {
	branchRef, err := s.GetBranchReference(ctx, branchName)
	if err == nil {
		return branchRef, nil
	}

	baseRef, err := s.GetDefaultBranchReference(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	newRef := github.Reference{
		Ref:    branchRef.Ref,
		Object: &github.GitObject{SHA: baseRef.Object.SHA},
	}
	branchRef, _, err = s.gitHubClient.Git.CreateRef(ctx, s.owner, s.repo, &newRef)
	return branchRef, microerror.Mask(err)
}

func (s *Client) GetWorktree(ctx context.Context, ref string) (billy.Filesystem, error) {
	fileSystem := memfs.New()
	_, err := git.CloneContext(ctx, memory.NewStorage(), fileSystem, &git.CloneOptions{
		Auth:          s.gitAuth,
		URL:           fmt.Sprintf("https://github.com/%s/%s.git", s.owner, s.repo),
		ReferenceName: plumbing.ReferenceName(ref),
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return fileSystem, nil
}

func (c Commit) toEntries() ([]*github.TreeEntry, error) {
	return nil, nil
}

// PushCommit creates the commit in the given reference using the given tree.
func (s *Client) PushCommit(ctx context.Context, branchName string, commit Commit) (*github.Reference, error) {
	branchRef, err := s.GetOrCreateBranch(ctx, branchName)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	commitSHA := branchRef.GetObject().GetSHA()
	parent, _, err := s.gitHubClient.Repositories.GetCommit(ctx, s.owner, s.repo, commitSHA, nil)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	parent.Commit.SHA = parent.SHA

	entries, err := commit.toEntries()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	tree, _, err := s.gitHubClient.Git.CreateTree(ctx, s.owner, s.repo, commitSHA, entries)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	date := time.Now()
	gitHubCommit := github.Commit{
		Author: &github.CommitAuthor{
			Date: &date,
			Name: github.String("capi-bootstrap"),
		},
		Message: github.String("update config"),
		Tree:    tree,
		Parents: []*github.Commit{parent.Commit},
	}
	newCommit, _, err := s.gitHubClient.Git.CreateCommit(ctx, s.owner, s.repo, &gitHubCommit)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	branchRef.Object.SHA = newCommit.SHA

	commitRef, _, err := s.gitHubClient.Git.UpdateRef(ctx, s.owner, s.repo, branchRef, false)
	return commitRef, microerror.Mask(err)
}

func (s *Client) UpdatePullRequest(ctx context.Context, pullRequest *github.PullRequest) (*github.PullRequest, error) {
	updated, _, err := s.gitHubClient.PullRequests.Edit(ctx, s.owner, s.repo, pullRequest.GetNumber(), pullRequest)
	return updated, microerror.Mask(err)
}

func headRefToBranchName(ref string) string {
	return strings.TrimPrefix(ref, "refs/heads/")
}

func branchNameToHeadRef(branch string) string {
	return fmt.Sprintf("refs/heads/%s", branch)
}

func (s *Client) defaultBranchName(ctx context.Context) (string, error) {
	defaultBranch, err := s.GetDefaultBranchReference(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return headRefToBranchName(defaultBranch.GetRef()), nil
}

func (s *Client) CreateOrUpdatePullRequest(ctx context.Context, pullRequest github.NewPullRequest) (*github.PullRequest, error) {
	createdPullRequest, _, err := s.gitHubClient.PullRequests.Create(ctx, s.owner, s.repo, &pullRequest)
	if ghErr, ok := err.(*github.ErrorResponse); ok {
		if len(ghErr.Errors) == 1 && strings.HasPrefix(ghErr.Errors[0].Message, "A pull request already exists for") {
			existing, err := s.GetPullRequest(ctx, pullRequest.GetHead())
			if err != nil {
				return nil, microerror.Mask(err)
			}
			if existing.Body == pullRequest.Body &&
				existing.Draft == pullRequest.Draft &&
				existing.Title == pullRequest.Title {
				return nil, nil
			}
			existing.Body = pullRequest.Body
			existing.Draft = pullRequest.Draft
			existing.Title = pullRequest.Title
			existing, err = s.UpdatePullRequest(ctx, existing)
			return existing, microerror.Mask(err)
		}
	}

	return createdPullRequest, microerror.Mask(err)
}

func (s *Client) GetPullRequest(ctx context.Context, branchName string) (*github.PullRequest, error) {
	matches, _, err := s.gitHubClient.PullRequests.List(ctx, s.owner, s.repo, &github.PullRequestListOptions{
		State: "open",
		Head:  branchName,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(matches) == 0 {
		return nil, microerror.Maskf(notFoundError, "pull request for branch %s not found", branchName)
	}

	return matches[0], nil
}

func (s *Client) ClosePullRequest(ctx context.Context, number int) error {
	_, _, err := s.gitHubClient.PullRequests.Edit(ctx, s.owner, s.repo, number, &github.PullRequest{
		State: github.String("closed"),
	})
	return microerror.Mask(err)
}
