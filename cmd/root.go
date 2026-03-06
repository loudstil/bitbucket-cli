package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"github.com/loudstil/bb/cmd/auth"
	"github.com/loudstil/bb/cmd/pr"
	"github.com/loudstil/bb/cmd/repo"
	"github.com/loudstil/bb/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "bb",
	Short: "Bitbucket CLI – interact with Cloud and Data Center from your terminal",
	Long: `bb is a unified CLI for Bitbucket Cloud and Bitbucket Data Center.

Get started by logging in:
  bb auth login

Then explore repositories and pull requests:
  bb repo list
  bb pr list`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(auth.NewAuthCmd())
	rootCmd.AddCommand(repo.NewRepoCmd())
	rootCmd.AddCommand(pr.NewPrCmd())
}

// Execute is the main entry point called by main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initConfig() {
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
