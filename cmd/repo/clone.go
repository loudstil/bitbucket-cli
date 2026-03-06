package repo

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/loudstil/bb/internal/config"
	"github.com/loudstil/bb/internal/factory"
	"github.com/loudstil/bb/pkg/api"
)

func newCloneCmd() *cobra.Command {
	var flagWorkspace string
	var flagProject string

	cmd := &cobra.Command{
		Use:   "clone <slug>",
		Short: "Clone a repository",
		Long: `Fetch the HTTPS clone URL for a repository from Bitbucket and run git clone.

For Cloud, the workspace is resolved in this order:
  1. --workspace flag
  2. Workspace stored in the active context (set during bb auth login)

For Data Center, --project is required.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClone(cmd, args[0], flagWorkspace, flagProject)
		},
	}

	cmd.Flags().StringVarP(&flagWorkspace, "workspace", "w", "", "Workspace slug (Cloud only; overrides context default)")
	cmd.Flags().StringVar(&flagProject, "project", "", "Project key (Data Center only; required)")

	return cmd
}

func runClone(cmd *cobra.Command, slug, flagWorkspace, flagProject string) error {
	client, err := factory.NewClient()
	if err != nil {
		return fmt.Errorf("repo clone: %w", err)
	}

	workspace, err := resolveWorkspace(client, flagWorkspace, flagProject)
	if err != nil {
		return fmt.Errorf("repo clone: %w", err)
	}

	repo, err := client.GetRepository(workspace, slug)
	if err != nil {
		return fmt.Errorf("repo clone: %w", err)
	}

	if repo.CloneURL == "" {
		return fmt.Errorf("repo clone: no HTTPS clone URL found for %s/%s", workspace, slug)
	}

	gitCmd := exec.Command("git", "clone", repo.CloneURL)
	gitCmd.Stdout = cmd.OutOrStdout()
	gitCmd.Stderr = cmd.ErrOrStderr()
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("repo clone: git clone failed: %w", err)
	}
	return nil
}

// resolveWorkspace returns the workspace/project key to use for API calls.
// For Cloud: --workspace flag → context.Workspace.
// For DC: --project flag (required).
func resolveWorkspace(client api.BitbucketClient, flagWorkspace, flagProject string) (string, error) {
	if client.ProviderType() == api.ProviderDataCenter {
		if flagProject == "" {
			return "", fmt.Errorf("--project is required for Data Center")
		}
		return flagProject, nil
	}

	// Cloud
	if flagWorkspace != "" {
		return flagWorkspace, nil
	}
	name := config.ActiveContext()
	ctx, err := config.GetContext(name)
	if err != nil {
		return "", err
	}
	if ctx.Workspace == "" {
		return "", fmt.Errorf("no workspace specified – use --workspace or re-run bb auth login")
	}
	return ctx.Workspace, nil
}
