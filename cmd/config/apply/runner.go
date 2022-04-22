package apply

import (
	"context"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/google/go-github/v43/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/config"
	"github.com/giantswarm/capi-bootstrap/pkg/repo"
)

func (r *Runner) Run(cmd *cobra.Command, args []string) error {
	err := r.flags.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context(), cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Runner) Do(ctx context.Context, _ *cobra.Command, _ []string) error {
	content, err := os.ReadFile(r.flags.ConfigFile)
	if err != nil {
		return microerror.Mask(err)
	}

	var bootstrapConfig config.ConfigFile
	err = yaml.Unmarshal(content, &bootstrapConfig)
	if err != nil {
		return microerror.Mask(err)
	}

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

	err = configService.EnsureCreated(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
