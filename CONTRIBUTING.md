# Contributing to kranix-cli

Thank you for your interest in contributing to kranix-cli!

## Development Setup

```bash
# Clone the repository
git clone https://github.com/kranix-io/kranix-cli
cd kranix-cli

# Install dependencies
go mod download

# Build the CLI
go build -o kranix ./cmd/kranix

# Run the CLI
./kranix --help
```

## Adding a New Command

Each new command needs:

1. **Cobra command definition** in `internal/commands/yourcommand.go`
2. **Help text** with `Short`, `Long`, and `Example` fields
3. **At least one example** in the command's usage
4. **Unit tests** for flag parsing in `tests/unit/`
5. **E2E test** against a running `kranix-api` in `tests/e2e/`

### Example Command Template

```go
package commands

import (
	"context"
	"fmt"

	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/kranix-io/kranix-cli/internal/output"
	"github.com/spf13/cobra"
)

var yourFlag string

var yourCmd = &cobra.Command{
	Use:   "your-command",
	Short: "Brief description",
	Long:  "Longer description with details about what the command does.",
	Example: `  kranix your-command --flag value`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, creds, err := getCredentials()
		if err != nil {
			return err
		}

		cli := client.New(creds.Server, creds.APIKey)
		// Your command logic here

		return nil
	},
}

func init() {
	yourCmd.Flags().StringVar(&yourFlag, "flag", "", "Flag description")
	rootCmd.AddCommand(yourCmd)
}
```

## Testing

### Unit Tests

```bash
go test ./internal/...
go test ./pkg/...
```

### E2E Tests

E2E tests require a running `kranix-api` instance:

```bash
# Start kranix-api (in another terminal)
cd ../kranix-api
go run ./cmd/api

# Run E2E tests
cd ../kranix-cli
go test ./tests/e2e/...
```

## Code Style

- Follow standard Go formatting: `go fmt ./...`
- Run linter: `golangci-lint run`
- Keep functions focused and small
- Use meaningful variable names
- Add comments for exported functions and complex logic

## Commit Messages

Follow conventional commits format:

- `feat: add new command for X`
- `fix: resolve issue with Y`
- `docs: update README for Z`
- `test: add unit tests for W`

## Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests
5. Ensure all tests pass
6. Commit your changes
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.
