package apply

import (
	"context"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/google/go-github/v43/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/giantswarm/capi-bootstrap/pkg/repo"
)

func (r *Runner) Run(cmd *cobra.Command, args []string) error {
	err := r.Do(cmd.Context(), cmd, args)
	return microerror.Mask(err)
}

func (r *Runner) Do(ctx context.Context, _ *cobra.Command, _ []string) error {
	var gitHubClient *github.Client
	{
		token := oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")}
		tokenSource := oauth2.StaticTokenSource(&token)
		httpClient := oauth2.NewClient(ctx, tokenSource)
		gitHubClient = github.NewClient(httpClient)
	}

	configService, err := repo.New(repo.Config{
		GitHubClient: gitHubClient,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	err = configService.EnsureDeleted(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
