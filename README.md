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

# Deploy a workload
kranix deploy --name my-app --image nginx:latest --namespace staging

# Watch status
kranix status my-app

# Stream logs
kranix logs my-app --follow

# Analyze a failing workload
kranix analyze my-app

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
  --name         Workload name (required)
  --image        Container image (required)
  --namespace    Target namespace (default: default)
  --replicas     Number of replicas (default: 1)
  --env          Environment variables (KEY=VALUE, repeatable)
  --port         Exposed port
  --cpu          CPU limit (e.g. 500m)
  --memory       Memory limit (e.g. 256Mi)
  --wait         Wait for workload to become ready
  --timeout      Timeout for --wait (default: 5m)
```

### `kranix status`

```
kranix status [workload-name] [flags]

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
