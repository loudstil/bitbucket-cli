package auth

import "github.com/spf13/cobra"

// NewAuthCmd returns the "auth" command group.
func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Bitbucket",
		Long:  "Manage authentication credentials for Bitbucket Cloud and Data Center.",
	}

	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newStatusCmd())

	return cmd
}
