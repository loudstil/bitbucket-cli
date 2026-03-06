// Package pr implements the "bb pr" command group.
package pr

import "github.com/spf13/cobra"

// NewPrCmd creates the "pr" command group and registers its sub-commands.
func NewPrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Manage Bitbucket pull requests",
	}
	cmd.AddCommand(newListCmd())
	return cmd
}
