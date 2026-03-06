// Package repo implements the "bb repo" command group.
package repo

import "github.com/spf13/cobra"

// NewRepoCmd creates the "repo" command group and registers its sub-commands.
func NewRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage Bitbucket repositories",
	}
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newCloneCmd())
	cmd.AddCommand(newCreateCmd())
	return cmd
}
