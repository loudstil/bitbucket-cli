package auth

import (
	"fmt"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/loudstil/bb/internal/config"
	"github.com/loudstil/bb/pkg/api"
	"github.com/loudstil/bb/pkg/api/cloud"
	bbkeyring "github.com/loudstil/bb/pkg/keyring"
)

func newLoginCmd() *cobra.Command {
	var flagToken    string
	var flagUsername string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a Bitbucket instance",
		Long: `Interactively log in to Bitbucket Cloud or a self-hosted Data Center instance.

Your API token is stored securely in the system keyring (Windows Credential
Manager, macOS Keychain, or the platform's secret service). It is never
written to the config file.

Flags --username and --token skip their respective prompts, which is useful
when the terminal corrupts pasted input or for scripted use:

  bb auth login --username you@example.com --token ATATT3x...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(cmd, flagUsername, flagToken)
		},
	}

	cmd.Flags().StringVarP(&flagUsername, "username", "u", "", "Email (Cloud) or username (DC) – skips the interactive prompt")
	cmd.Flags().StringVarP(&flagToken, "token", "t", "", "API token – skips the interactive prompt")

	return cmd
}

func runLogin(cmd *cobra.Command, flagUsername, flagToken string) error {
	// ------------------------------------------------------------------ //
	// Step 1: Choose provider type
	// ------------------------------------------------------------------ //
	var providerChoice string
	if err := survey.AskOne(&survey.Select{
		Message: "Which Bitbucket product are you logging in to?",
		Options: []string{"Bitbucket Cloud", "Bitbucket Data Center / Server"},
	}, &providerChoice); err != nil {
		return fmt.Errorf("auth login: provider selection: %w", err)
	}

	isCloud := strings.HasPrefix(providerChoice, "Bitbucket Cloud")
	providerType := config.ProviderCloud
	baseURL := "https://api.bitbucket.org/2.0"

	// ------------------------------------------------------------------ //
	// Step 2: Base URL (Data Center only)
	// ------------------------------------------------------------------ //
	if !isCloud {
		providerType = config.ProviderDC
		if err := survey.AskOne(&survey.Input{
			Message: "Enter your Data Center base URL (e.g. https://bitbucket.example.com):",
			Help:    "Do not include /rest/api/1.0 – bb appends that automatically.",
		}, &baseURL, survey.WithValidator(survey.Required)); err != nil {
			return fmt.Errorf("auth login: base URL: %w", err)
		}
		baseURL = strings.TrimRight(baseURL, "/")
	}

	// ------------------------------------------------------------------ //
	// Step 3: Context / profile name
	// ------------------------------------------------------------------ //
	defaultName := "cloud"
	if !isCloud {
		defaultName = hostFromURL(baseURL)
	}

	var contextName string
	if err := survey.AskOne(&survey.Input{
		Message: "Give this connection a name (used as a profile alias):",
		Default: defaultName,
		Help:    "You can switch between profiles with: bb auth switch <name>",
	}, &contextName, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("auth login: context name: %w", err)
	}
	contextName = strings.TrimSpace(contextName)

	// ------------------------------------------------------------------ //
	// Step 4: Email / username  (use flag if provided)
	// ------------------------------------------------------------------ //
	username := strings.TrimSpace(flagUsername)
	if username == "" {
		emailPrompt := "Your Bitbucket account email:"
		if !isCloud {
			emailPrompt = "Your Bitbucket username:"
		}
		if err := survey.AskOne(&survey.Input{
			Message: emailPrompt,
		}, &username, survey.WithValidator(survey.Required)); err != nil {
			return fmt.Errorf("auth login: username: %w", err)
		}
		username = strings.TrimSpace(username)
	}

	// ------------------------------------------------------------------ //
	// Step 5: API token  (use flag if provided, otherwise prompt)
	// ------------------------------------------------------------------ //
	token := strings.TrimSpace(flagToken)
	if token == "" {
		tokenHelp := "Create a Cloud API token at: https://id.atlassian.com/manage-profile/security/api-tokens"
		if !isCloud {
			tokenHelp = "Create a DC personal access token under your profile → Manage tokens."
		}
		if err := survey.AskOne(&survey.Password{
			Message: "Paste your API token:",
			Help:    tokenHelp,
		}, &token, survey.WithValidator(survey.Required)); err != nil {
			return fmt.Errorf("auth login: token input: %w", err)
		}
		token = strings.TrimSpace(token)
	}

	// Strip any non-printable / control characters that a terminal might
	// silently inject during paste (common in Windows Terminal and ConHost).
	token = stripControlChars(token)

	// ------------------------------------------------------------------ //
	// Step 6: Verify credentials against the live API before saving anything
	// ------------------------------------------------------------------ //
	fmt.Fprintf(cmd.OutOrStdout(), "\nVerifying credentials...")
	verified, err := api.VerifyCredentials(api.ProviderType(providerType), baseURL, username, token)
	if err != nil {
		fmt.Fprintln(cmd.OutOrStdout(), " failed.")
		return fmt.Errorf("auth login: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), " OK")

	// For Cloud, trust the identity returned by the API.
	// Keep the original email — Basic Auth always needs it, even after we
	// swap `username` to the nickname for display/storage purposes.
	email := username
	if providerType == config.ProviderCloud && verified.Username != "" {
		username = verified.Username
	}

	// ------------------------------------------------------------------ //
	// Step 6b: Workspace slug (Cloud only)
	// ------------------------------------------------------------------ //
	var workspace string
	if isCloud {
		workspace, err = promptWorkspace(cmd, email, token)
		if err != nil {
			return fmt.Errorf("auth login: workspace: %w", err)
		}
	}

	// ------------------------------------------------------------------ //
	// Step 7: Persist – token → keyring, metadata → config file
	// ------------------------------------------------------------------ //
	if err := bbkeyring.Set(contextName, token); err != nil {
		return fmt.Errorf("auth login: save token: %w", err)
	}

	ctx := config.Context{
		Name:      contextName,
		Type:      providerType,
		BaseURL:   baseURL,
		Username:  email, // Cloud: email for Basic Auth; DC: username (same as email here)
		Workspace: workspace,
	}
	if err := config.AddContext(ctx); err != nil {
		return fmt.Errorf("auth login: save context: %w", err)
	}
	if err := config.SetActiveContext(contextName); err != nil {
		return fmt.Errorf("auth login: set active context: %w", err)
	}

	displayName := username // nickname for Cloud, DC username otherwise
	if providerType == config.ProviderCloud && verified.DisplayName != "" {
		displayName = verified.DisplayName
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s on context %q.\n", displayName, contextName)
	fmt.Fprintln(cmd.OutOrStdout(), "Token stored securely in the system keyring.")
	return nil
}

// promptWorkspace fetches the user's available workspaces and presents a select
// prompt. If the API call fails or returns no results it falls back to a free-
// text input so login can still complete without network access.
func promptWorkspace(cmd *cobra.Command, email, token string) (string, error) {
	fmt.Fprint(cmd.OutOrStdout(), "Fetching workspaces...")
	workspaces, err := cloud.ListWorkspaces(email, token)
	fmt.Fprintln(cmd.OutOrStdout())

	if err == nil && len(workspaces) > 0 {
		// Build display options: "name (slug)" so the user sees both.
		options := make([]string, len(workspaces))
		slugByOption := make(map[string]string, len(workspaces))
		for i, w := range workspaces {
			label := w.Slug
			if w.Name != "" && w.Name != w.Slug {
				label = w.Name + " (" + w.Slug + ")"
			}
			options[i] = label
			slugByOption[label] = w.Slug
		}

		var chosen string
		if err := survey.AskOne(&survey.Select{
			Message: "Select a workspace:",
			Options: options,
		}, &chosen, survey.WithValidator(survey.Required)); err != nil {
			return "", err
		}
		return slugByOption[chosen], nil
	}

	// Fallback: manual input.
	if err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Could not fetch workspaces (%v). Enter slug manually.\n", err)
	}
	var slug string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter your Bitbucket workspace slug:",
		Help:    "Found in the URL: bitbucket.org/{workspace}/",
	}, &slug, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}
	return strings.TrimSpace(slug), nil
}

// stripControlChars removes ASCII control characters (0x00–0x1F, 0x7F) from s.
// This guards against terminals silently injecting escape sequences or carriage
// returns when the user pastes a token into a masked prompt.
func stripControlChars(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7F {
			return -1 // drop the character
		}
		return r
	}, s)
}

// hostFromURL extracts a short hostname from a URL for use as a default
// profile name (e.g. "https://bitbucket.example.com" → "bitbucket.example.com").
func hostFromURL(rawURL string) string {
	u := strings.TrimPrefix(rawURL, "https://")
	u = strings.TrimPrefix(u, "http://")
	if idx := strings.Index(u, "/"); idx != -1 {
		u = u[:idx]
	}
	if u == "" {
		return "datacenter"
	}
	return u
}
