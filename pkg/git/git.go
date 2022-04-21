package git

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/google/go-github/v43/github"
)

func New(config Config) (*Client, error) {
	return &Client{
		gitHubClient: config.GitHubClient,

		owner: config.Owner,
		repo:  config.Repo,

		branchName:     config.BranchName,
		commitAuthor:   config.CommitAuthor,
		commitTemplate: config.CommitTemplate,
	}, nil
}

func (c *Client) GetDefaultBranch(ctx context.Context) (*github.Reference, error) {
	if c.defaultBranch == nil {
		repo, _, err := c.gitHubClient.Repositories.Get(ctx, c.owner, c.repo)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		refName := branchNameToHeadRef(repo.GetDefaultBranch())
		c.defaultBranch, _, err = c.gitHubClient.Git.GetRef(ctx, c.owner, c.repo, refName)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return c.defaultBranch, nil
}

func (c *Client) GetOrCreateBranch(ctx context.Context) (*github.Reference, error) {
	branchRefName := fmt.Sprintf("refs/heads/%s", c.branchName)
	branchRef, _, err := c.gitHubClient.Git.GetRef(ctx, c.owner, c.repo, branchRefName)
	if err == nil {
		return branchRef, nil
	}

	baseRef, err := c.GetDefaultBranch(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	newRef := github.Reference{
		Ref:    github.String(branchRefName),
		Object: &github.GitObject{SHA: baseRef.Object.SHA},
	}
	branchRef, _, err = c.gitHubClient.Git.CreateRef(ctx, c.owner, c.repo, &newRef)
	return branchRef, err
}

// PushCommit creates the commit in the given reference using the given tree.
func (c *Client) PushCommit(ctx context.Context, entries []*github.TreeEntry) (*github.Reference, error) {
	branchRef, err := c.GetOrCreateBranch(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	parent, _, err := c.gitHubClient.Repositories.GetCommit(ctx, c.owner, c.repo, branchRef.GetObject().GetSHA(), nil)
	if err != nil {
		return nil, err
	}

	parent.Commit.SHA = parent.SHA

	tree, _, err := c.gitHubClient.Git.CreateTree(ctx, c.owner, c.repo, branchRef.GetObject().GetSHA(), entries)
	if err != nil {
		return nil, err
	}

	date := time.Now()
	author := c.commitAuthor
	author.Date = &date
	commit := github.Commit{
		Author:  &author,
		Message: c.commitTemplate.Message,
		Tree:    tree,
		Parents: []*github.Commit{parent.Commit},
	}
	newCommit, _, err := c.gitHubClient.Git.CreateCommit(ctx, c.owner, c.repo, &commit)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	branchRef.Object.SHA = newCommit.SHA

	commitRef, _, err := c.gitHubClient.Git.UpdateRef(ctx, c.owner, c.repo, branchRef, false)
	return commitRef, microerror.Mask(err)
}

func (c *Client) UpdatePullRequest(ctx context.Context, pullRequest github.NewPullRequest) (*github.PullRequest, error) {
	existingPulls, _, err := c.gitHubClient.PullRequests.List(ctx, c.owner, c.repo, &github.PullRequestListOptions{
		Head: c.branchName,
		Base: pullRequest.GetBase(),
	})
	if err != nil {
		return nil, microerror.Mask(err)
	} else if len(existingPulls) == 0 {
		return nil, errors.New("pr not found")
	}

	existingPull := existingPulls[0]
	var shouldUpdate bool
	if existingPull.Body != pullRequest.Body ||
		existingPull.Title != pullRequest.Title ||
		existingPull.Draft != pullRequest.Draft ||
		existingPull.MaintainerCanModify != pullRequest.MaintainerCanModify {
		shouldUpdate = true
	}
	if !shouldUpdate {
		return existingPull, nil
	}

	existingPull.Body = pullRequest.Body
	existingPull.Title = pullRequest.Title
	existingPull.Draft = pullRequest.Draft
	existingPull.MaintainerCanModify = pullRequest.MaintainerCanModify

	existingPull, _, err = c.gitHubClient.PullRequests.Edit(ctx, c.owner, c.repo, existingPull.GetNumber(), existingPull)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return existingPull, nil
}

func headRefToBranchName(ref string) string {
	return strings.TrimPrefix(ref, "refs/heads/")
}

func branchNameToHeadRef(branch string) string {
	return fmt.Sprintf("refs/heads/%s", branch)
}

func (c *Client) defaultBranchName(ctx context.Context) (string, error) {
	defaultBranch, err := c.GetDefaultBranch(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return headRefToBranchName(defaultBranch.GetRef()), nil
}

func (c *Client) CreateOrUpdatePullRequest(ctx context.Context, pullRequest github.NewPullRequest) (*github.PullRequest, error) {
	{
		branchRef, err := c.GetOrCreateBranch(ctx)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		pullRequest.Head = github.String(headRefToBranchName(branchRef.GetRef()))
	}

	{
		baseBranchName, err := c.defaultBranchName(ctx)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		pullRequest.Base = github.String(baseBranchName)
	}

	createdPullRequest, _, err := c.gitHubClient.PullRequests.Create(ctx, c.owner, c.repo, &pullRequest)
	if ghErr, ok := err.(*github.ErrorResponse); ok {
		if len(ghErr.Errors) == 1 && strings.HasPrefix(ghErr.Errors[0].Message, "A pull request already exists for") {
			existing, err := c.UpdatePullRequest(ctx, pullRequest)
			return existing, microerror.Mask(err)
		}
	}

	return createdPullRequest, microerror.Mask(err)
}
