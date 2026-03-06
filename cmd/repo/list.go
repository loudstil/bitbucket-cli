package repo

import (
	"encoding/json"
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"

	"github.com/loudstil/bb/internal/config"
	"github.com/loudstil/bb/internal/factory"
	"github.com/loudstil/bb/pkg/api"
)

func newListCmd() *cobra.Command {
	var flagWorkspace string
	var flagJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories",
		Long: `List repositories in a Bitbucket workspace (Cloud) or all accessible
repositories (Data Center).

For Cloud, the workspace is resolved in this order:
  1. --workspace flag
  2. Workspace stored in the active context (set during bb auth login)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, flagWorkspace, flagJSON)
		},
	}

	cmd.Flags().StringVarP(&flagWorkspace, "workspace", "w", "", "Workspace slug (Cloud only; overrides context default)")
	cmd.Flags().BoolVar(&flagJSON, "json", false, "Output as JSON array")

	return cmd
}

func runList(cmd *cobra.Command, flagWorkspace string, jsonOut bool) error {
	client, err := factory.NewClient()
	if err != nil {
		return fmt.Errorf("repo list: %w", err)
	}

	// For Cloud, resolve workspace: --workspace flag > context.Workspace.
	workspace := flagWorkspace
	if workspace == "" && client.ProviderType() == api.ProviderCloud {
		name := config.ActiveContext()
		ctx, err := config.GetContext(name)
		if err != nil {
			return fmt.Errorf("repo list: %w", err)
		}
		workspace = ctx.Workspace
		if workspace == "" {
			return fmt.Errorf("repo list: no workspace specified – use --workspace or re-run bb auth login")
		}
	}

	repos, err := client.ListRepositories(workspace)
	if err != nil {
		return fmt.Errorf("repo list: %w", err)
	}

	if jsonOut {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(repos)
	}

	printTable(cmd, repos)
	return nil
}

const maxDesc = 60

func printTable(cmd *cobra.Command, repos []api.Repository) {
	table := tablewriter.NewTable(cmd.OutOrStdout(),
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Borders: tw.BorderNone,
			Symbols: tw.NewSymbols(tw.StyleASCII),
			Settings: tw.Settings{
				Lines: tw.LinesNone,
			},
		})),
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
		tablewriter.WithRowAlignment(tw.AlignLeft),
	)

	table.Header([]string{"SLUG", "FULL NAME", "PRIVATE", "DESCRIPTION"})

	for _, r := range repos {
		private := "false"
		if r.IsPrivate {
			private = "true"
		}
		desc := r.Description
		if len([]rune(desc)) > maxDesc {
			desc = string([]rune(desc)[:maxDesc-3]) + "..."
		}
		table.Append([]string{r.Slug, r.FullName, private, desc})
	}
	table.Render()
}
