package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/loudstil/bb/internal/config"
	bbkeyring "github.com/loudstil/bb/pkg/keyring"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status for all saved contexts",
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, _ []string) error {
	contexts, err := config.ListContexts()
	if err != nil {
		return fmt.Errorf("auth status: %w", err)
	}

	if len(contexts) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No contexts configured. Run: bb auth login")
		return nil
	}

	active := config.ActiveContext()
	out := cmd.OutOrStdout()

	for _, ctx := range contexts {
		marker := "  "
		if ctx.Name == active {
			marker = "* "
		}

		// Verify a token actually exists in the keyring (don't print it).
		tokenOK := "token: OK"
		if _, err := bbkeyring.Get(ctx.Name); err != nil {
			tokenOK = "token: MISSING (re-run bb auth login)"
		}

		displayURL := ctx.BaseURL
		if ctx.Type == config.ProviderCloud {
			displayURL = "https://bitbucket.org"
		}

		fmt.Fprintf(out, "%s%s  %s  user:%s  %s  %s\n",
			marker, ctx.Name, ctx.Type, ctx.Username, displayURL, tokenOK)
	}
	return nil
}
