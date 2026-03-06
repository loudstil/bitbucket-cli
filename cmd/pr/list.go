package pr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"

	"github.com/loudstil/bb/internal/config"
	"github.com/loudstil/bb/internal/factory"
	"github.com/loudstil/bb/pkg/api"
	"github.com/loudstil/bb/pkg/git"
)

func newListCmd() *cobra.Command {
	var flagWorkspace string
	var flagRepo     string
	var flagState    string
	var flagJSON     bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pull requests for a repository",
		Long: `List pull requests for a Bitbucket repository.

The repository is auto-detected from the git remote URL of the current
directory. Use --workspace and/or --repo to override or specify explicitly.

State filter (--state): OPEN (default), MERGED, DECLINED, ALL.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPRList(cmd, flagWorkspace, flagRepo, strings.ToUpper(flagState), flagJSON)
		},
	}

	cmd.Flags().StringVarP(&flagWorkspace, "workspace", "w", "", "Workspace/project key (overrides git detection)")
	cmd.Flags().StringVarP(&flagRepo, "repo", "r", "", "Repository slug (overrides git detection)")
	cmd.Flags().StringVar(&flagState, "state", "OPEN", "Filter by state: OPEN, MERGED, DECLINED, ALL")
	cmd.Flags().BoolVar(&flagJSON, "json", false, "Output as JSON array")

	return cmd
}

func runPRList(cmd *cobra.Command, flagWorkspace, flagRepo, state string, jsonOut bool) error {
	switch state {
	case "OPEN", "MERGED", "DECLINED", "ALL":
	default:
		return fmt.Errorf("pr list: invalid state %q – use OPEN, MERGED, DECLINED, or ALL", state)
	}

	client, err := factory.NewClient()
	if err != nil {
		return fmt.Errorf("pr list: %w", err)
	}

	workspace := flagWorkspace
	slug := flagRepo

	// Try git detection when either value is still missing.
	if workspace == "" || slug == "" {
		info, gitErr := git.Detect()
		if gitErr == nil {
			if workspace == "" {
				workspace = info.Workspace
			}
			if slug == "" {
				slug = info.Slug
			}
		}
		// If git fails we fall through; missing values are caught below.
	}

	// For Cloud, fall back to the context workspace when still empty.
	if workspace == "" && client.ProviderType() == api.ProviderCloud {
		name := config.ActiveContext()
		ctx, err := config.GetContext(name)
		if err != nil {
			return fmt.Errorf("pr list: %w", err)
		}
		workspace = ctx.Workspace
	}

	if slug == "" {
		return fmt.Errorf("pr list: repository slug is required – use --repo or run from a git repo cloned from Bitbucket")
	}
	if workspace == "" {
		return fmt.Errorf("pr list: workspace/project key is required – use --workspace or run from a git repo")
	}

	prs, err := client.ListPullRequests(workspace, slug, state)
	if err != nil {
		return fmt.Errorf("pr list: %w", err)
	}

	if jsonOut {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(prs)
	}

	printPRTable(cmd, prs)
	return nil
}

const maxTitle = 60

func printPRTable(cmd *cobra.Command, prs []api.PullRequest) {
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

	table.Header([]string{"ID", "TITLE", "AUTHOR", "BRANCHES", "STATE"})

	for _, pr := range prs {
		title := pr.Title
		if len([]rune(title)) > maxTitle {
			title = string([]rune(title)[:maxTitle-3]) + "..."
		}
		branches := pr.SourceBranch + " → " + pr.TargetBranch
		table.Append([]string{
			fmt.Sprintf("%d", pr.ID),
			title,
			pr.AuthorName,
			branches,
			pr.State,
		})
	}
	table.Render()
}
