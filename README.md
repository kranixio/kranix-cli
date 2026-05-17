# kranix-cli

> Terminal UX — deploy, inspect, and debug your infrastructure from the shell.

`kranix-cli` is the primary command-line interface for the Kranix platform. It wraps `kranix-api` in a fast, developer-friendly shell experience with commands for deploying workloads, streaming logs, inspecting cluster state, analyzing failures, and managing namespaces. If `kranix-mcp` is how AI agents operate Kranix, `kranix-cli` is how humans do.

---

## What it does

- Provides a clean `kranix <command>` CLI for all common platform operations
- Authenticates with `kranix-api` using API keys or JWT (via `kranix login`)
- Streams logs and events in real time to the terminal
- Outputs in human-readable tables or `--json` / `--yaml` for scripting
- Supports shell completion for bash, zsh, and fish
- Stores context (server URL + credentials) in `~/.kranix/config`

---

## Architecture position

```
Developer (terminal)
       │
   kranix-cli  ──►  kranix-api  ──►  kranix-core
```

`kranix-cli` is a pure API client. All business logic lives in `kranix-api` and `kranix-core`. The CLI is responsible only for UX: input parsing, output formatting, and streaming.

---

## Installation

### Homebrew (macOS / Linux)

```bash
brew install kranix-io/tap/kranix
```

### curl installer

```bash
curl -fsSL https://get.kranix.io | sh
```

### Go install

```bash
go install github.com/kranix-io/kranix-cli/cmd/kranix@latest
```

### From source

```bash
git clone https://github.com/kranix-io/kranix-cli
cd kranix-cli
go build -o kranix ./cmd/kranix
mv kranix /usr/local/bin/kranix
```

---

## Quick start

```bash
# Authenticate with your Kranix API server
kranix login --server http://localhost:8080 --api-key krane_your_key

# Deploy a workload tatus
kranix status my-appnamepace saging --st

# Stream logsmuti-clst fdion
kranix logs my-app --followfedeion-fdratio-clus ersaciusner east,workloa-wst

# Cee and manage ppelies
kranix pipnlinee mfilppielne.yal
kr pipeinlist
knix pipelenam plypipein-ppov aging

# List everything in a namespace
kranix list --namespace staging
```

---

## Command reference

### `kranix login`

```
kranix login [flags]

Flags:
  --server      Kranix API server URL (default: http://localhost:8080)
  --api-key     API key for authentication
  --oidc        Use OIDC browser-based login
```

### `kranix deploy`

```
kranix deploy [flags]

Flags:
  --name         me (required)
  -   Container image (required)
ce    Target namespace (default: default)
  --replica Numbeeplicas (default: 1)
      Environment variables (KEY=VALUE, repeatable)
  -   Exposed port
  --cp  CPU limite.g. 500m)
      Memory limit (e.g. 256Mi)
  --waWait for workload to become ready
  --timemeout for --wait (default: 5m)

Flags:
  --namespace    Filter by namespace
  --watch        Watch for status changes
  --json         Output as JSON
```

### `kranix logs`

```
kranix logs <workload-name> [flags]

Flags:
  --follow       Stream logs in real time (-f)
  --tail         Number of lines to show from end (default: 100)
  --since        Show logs since a duration (e.g. 10m, 1h)
  --pod          Target a specific pod
  --json         Output as JSON
```

### `kranix analyze`

```
kranix analyze <workload-name>

Runs AI-powered failure analysis via kranix-api. Returns:
  - Crash reason
  - Probable fix
  - Resource bottleneck detection
  - Failing dependency identification
  - Generated patch (if applicable)
```

### `kranix restart`

```
kranix restart <workload-name> [flags]

Flags:
  --namespace    Target namespace
```

### `kranix delete`

```
kranix delete <workload-name> [flags]

Flags:
  --namespace    Target namespace
  --confirm      Skip interactive confirmation prompt
```

### `kranix namespace`

```
kranix namespace create <name>
kranix namespace list
kranix namespace delete <name>
```

### `kranix manifests generate`

```
kranix manifests generate [flags]

Flags:
  --from         Plain-text description of what you want to deploy
  --output       Output file path (default: stdout)
  --format       yaml | json (default: yaml)

Example:
  kranix manifests generate --from "a redis instance with 1 replica and persistent storage"
```

### `kranix context`

```
kranix context list                  # list saved server contexts
kranix context use <name>            # switch active context
kranix context set-default <name>    # set default context
kranix context profile set <name> <server> [--api-key <key>]  # set or update a profile
kranix context profile switch <name> # switch to a profile (dev, staging, prod)
```

### `kranix doctor`

```
kranix doctor

Flags:
  --verbose    Show detailed diagnostic information
  --fix        Attempt to auto-fix issues

Diagnose local setup issues and config problems. Checks config file,
context configuration, API connectivity, credentials, required tools,
and directory permissions.
```

### `kranix port-forward`

```
kranix port-forward <workload-name> [flags]

Flags:
  --namespace    Target namespace
  --local, -l    Local port (default: 8080)
  --remote, -r   Remote pod port (default: 80)
  --pod, -p      Specific pod name (auto-selects if not specified)

Shorthand for forwarding a local port to any workload pod. Automatically
selects a pod and forwards the specified port. Requires kubectl to be installed.
```

### `kranix dashboard`

```
kranix dashboard

Flags:
  --namespace    Filter by namespace
  --refresh      Refresh interval in seconds (default: 5)

Launch a k9s-style TUI dashboard for real-time cluster state monitoring.
Navigate with arrow keys or vim-style (j/k), press enter for details, q to quit.
```

### `kranix diff`

```
kranix diff [workload-name] [flags]

Flags:
  --namespace    Target namespace
  --output       Output format: table, json, yaml (default: table)
  --file         Diff from manifest file
  --image        New container image
  --replicas     New replica count
  --env          Environment variables (KEY=VALUE)

Show exactly what would change before applying. Uses dry-run mode to preview changes.
```

### `kranix ai`

```
kranix ai [flags]

Flags:
  --namespace    Target namespace
  --context      Provide context for one-shot analysis
  --mode         Mode: chat, analyze, fix (default: chat)

Launch an inline AI assistant that analyzes your context, diagnoses issues,
and suggests fixes directly in the terminal. Supports interactive chat mode
or one-shot analysis with --context.
```

### `kranix cost`

```
kranix cost [workload-name] [flags]

Flags:
  --namespace    Filter by namespace
  --output       Output format: table, json, yaml (default: table)
  --duration     Time duration (e.g., 1h, 1d, 7d, 30d) (default: 7d)

Show estimated spend breakdown per workload, including compute, storage,
and network costs. View cost for a specific workload or get a summary
for all workloads in a namespace.
```

### `kranix template`

```
kranix template [flags]

Flags:
  --list         List available templates
  --name         Template name to apply
  --output       Output file path (default: stdout)
  --var          Template variables (KEY=VALUE)

Bootstrap a new workload from community templates. List available templates
or apply a template with custom variables. Supports popular templates like
nginx, nodejs, postgres, and more.
```

### `kranix offline`

```
kranix offline [flags]

Flags:
  --enable       Enable offline mode
  --disable      Disable offline mode
  --status       Show offline mode status
  --sync         Sync offline cache with API

Manage offline/air-gap mode for basic operations when kranix-api is not
reachable. Caches workloads locally for read-only access. Enable offline mode
and sync cache to work without API connectivity.
```

### `kranix pipeline`

```
kranix pipeline [flags]

Flags:
  --name, -n          Pipeline name
  --file, -f          Pipeline manifest file
  --approve           Approve a specific stage
  --list, -l          List all pipelines

Manage multi-stage deployment pipelines. Create pipelines from manifest files,
list all pipelines, or approve manual approval gates.

Examples:
  kranix pipeline --file pipeline.yaml
  kranix pipeline --list
  kranix pipeline --name deploy-pipeline --approve staging
```

---

## Shell completion

```bash
# bash
kranix completion bash > /etc/bash_completion.d/kranix

# zsh
kranix completion zsh > "${fpath[1]}/_krane"

# fish
kranix completion fish > ~/.config/fish/completions/kranix.fish
```

---

## Configuration

Config file lives at `~/.kranix/config`:

```yaml
current-context: local

contexts:
  - name: local
    server: http://localhost:8080
    api-key: krane_abc123

  - name: production
    server: https://kranix.mycompany.com
    api-key: krane_xyz789

defaults:
  namespace: default
  output: table        # table | json | yaml
  timeout: 5m
```

Override any value with environment variables:

```bash
KRANE_SERVER=http://localhost:8080
KRANE_API_KEY=krane_abc123
KRANE_NAMESPACE=staging
```

---

## Project structure

```
kranix-cli/
├── cmd/
│   └── kranix/            # Entry point (cobra root command)
├── internal/
│   ├── commands/         # One file per command (deploy.go, logs.go, etc.)
│   ├── client/           # kranix-api HTTP client
│   ├── output/           # Table, JSON, YAML formatters
│   ├── auth/             # Login, token storage, refresh
│   └── config/           # Config file read/write
├── pkg/
│   └── completion/       # Shell completion helpers
└── tests/
    ├── unit/
    └── e2e/              # Requires running kranix-api
```

---

## Connectivity

| Repo | Relationship |
|---|---|
| `kranix-api` | All commands translate into kranix-api HTTP requests |
| `kranix-packages` | Imports shared API types and error codes |

---

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md). Each new command needs: Cobra command definition, help text, at least one example, unit tests for flag parsing, and an E2E test against a running `kranix-api`.

## License

Apache 2.0 — see [LICENSE](./LICENSE).
