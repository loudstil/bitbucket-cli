# bb — Bitbucket CLI

A unified command-line interface for **Bitbucket Cloud** and **Bitbucket Data Center**.

## Installation

```bash
git clone https://github.com/loudstil/bb
cd bb
go build -o bb .
```

Move the binary somewhere on your `PATH`, e.g. `/usr/local/bin/bb` (Linux/macOS) or add its directory to `PATH` on Windows.

---

## Getting started

```bash
bb auth login
bb repo list
bb pr list
```

---

## Commands

### `bb auth`

Manage authentication credentials.

#### `bb auth login`

Interactively log in to Bitbucket Cloud or a self-hosted Data Center instance.
Prompts for provider type, credentials, and workspace (Cloud) or base URL (DC).
The API token is stored in the system keyring (Windows Credential Manager, macOS Keychain, or the platform secret service) and never written to disk.

```
bb auth login [flags]

Flags:
  -u, --username STRING   Email (Cloud) or username (DC) — skips the interactive prompt
  -t, --token    STRING   API token — skips the interactive prompt
```

**Cloud token:** generate at https://id.atlassian.com/manage-profile/security/api-tokens
**Data Center token:** profile → Manage tokens

#### `bb auth status`

Show authentication status for all saved profiles.

```
bb auth status
```

Output shows the active profile (marked with `*`), provider type, username, base URL, and whether a token is present in the keyring.

---

### `bb repo`

Manage Bitbucket repositories.

#### `bb repo list`

List repositories in a Cloud workspace or all accessible repositories on Data Center.

```
bb repo list [flags]

Flags:
  -w, --workspace STRING   Cloud workspace slug (overrides the context default)
      --json               Output as a JSON array
```

#### `bb repo clone <slug>`

Look up the HTTPS clone URL for a repository and run `git clone`.

```
bb repo clone <slug> [flags]

Flags:
  -w, --workspace STRING   Cloud workspace slug (overrides the context default)
      --project   STRING   Data Center project key (required for DC)
```

**Workspace resolution (Cloud):** `--workspace` flag → workspace stored in the active context.

#### `bb repo create <slug>`

Create a new repository on Cloud or Data Center.

```
bb repo create <slug> [flags]

Flags:
  -w, --workspace   STRING   Cloud workspace slug (overrides the context default)
      --project     STRING   Data Center project key (required for DC)
      --description STRING   Repository description
      --private              Make the repository private (Cloud only; DC access is project-level)
```

Prints the web URL of the newly created repository on success.

---

### `bb pr`

Manage Bitbucket pull requests.

#### `bb pr list`

List pull requests for a repository.

```
bb pr list [flags]

Flags:
  -w, --workspace STRING   Workspace or project key (overrides git detection)
  -r, --repo      STRING   Repository slug (overrides git detection)
      --state     STRING   Filter by state: OPEN, MERGED, DECLINED, ALL (default: OPEN)
      --json               Output as a JSON array
```

The workspace and repository slug are auto-detected from the `origin` remote of the current git directory. Use `--workspace` and `--repo` to override or when running outside a git repo.

---

## Configuration

Config is stored at `~/.config/bb/config.yaml`. Tokens are **never** written there — they live in the system keyring only.

Multiple profiles (contexts) are supported. The active profile is set automatically on `bb auth login` and shown with a `*` in `bb auth status`.

---

## Supported platforms

| OS      | Architecture |
|---------|-------------|
| Linux   | amd64, arm64 |
| macOS   | amd64, arm64 |
| Windows | amd64 |
