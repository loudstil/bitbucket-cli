package repo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/loudstil/bb/internal/factory"
	"github.com/loudstil/bb/pkg/api"
)

func newCreateCmd() *cobra.Command {
	var flagWorkspace string
	var flagProject string
	var flagDescription string
	var flagPrivate bool

	cmd := &cobra.Command{
		Use:   "create <slug>",
		Short: "Create a new repository",
		Long: `Create a new repository on Bitbucket Cloud or Data Center.

For Cloud, the workspace is resolved in this order:
  1. --workspace flag
  2. Workspace stored in the active context (set during bb auth login)

For Data Center, --project is required. The --private flag is ignored for DC
since access control is project-level.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args[0], flagWorkspace, flagProject, flagDescription, flagPrivate)
		},
	}

	cmd.Flags().StringVarP(&flagWorkspace, "workspace", "w", "", "Workspace slug (Cloud only; overrides context default)")
	cmd.Flags().StringVar(&flagProject, "project", "", "Project key (Data Center only; required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Repository description")
	cmd.Flags().BoolVar(&flagPrivate, "private", false, "Make repository private (Cloud only)")

	return cmd
}

func runCreate(cmd *cobra.Command, slug, flagWorkspace, flagProject, description string, private bool) error {
	client, err := factory.NewClient()
	if err != nil {
		return fmt.Errorf("repo create: %w", err)
	}

	workspace, err := resolveWorkspace(client, flagWorkspace, flagProject)
	if err != nil {
		return fmt.Errorf("repo create: %w", err)
	}

	created, err := client.CreateRepository(workspace, slug, api.CreateRepoRequest{
		Description: description,
		IsPrivate:   private,
	})
	if err != nil {
		return fmt.Errorf("repo create: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created repository: %s\n", created.WebURL)
	return nil
}
