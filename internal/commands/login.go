package commands

import (
	"fmt"
	"os"

	"github.com/kranix-io/kranix-cli/internal/auth"
	"github.com/kranix-io/kranix-cli/internal/config"
	"github.com/spf13/cobra"
)

var loginServer string
var loginAPIKey string
var loginOIDC bool

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Kranix API server",
	Long:  "Authenticate with your Kranix API server using API keys or OIDC.",
	Example: `  kranix login --server http://localhost:8080 --api-key krane_abc123
  kranix login --oidc`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if loginOIDC {
			return fmt.Errorf("OIDC login not yet implemented")
		}

		if loginAPIKey == "" {
			return fmt.Errorf("--api-key is required (or set KRANE_API_KEY env var)")
		}

		if err := auth.ValidateAPIKey(loginAPIKey); err != nil {
			return fmt.Errorf("invalid API key: %w", err)
		}

		server := loginServer
		if server == "" {
			if server = os.Getenv("KRANE_SERVER"); server == "" {
				server = "http://localhost:8080"
			}
		}

		// Update or add context
		ctxName := "default"
		if server != "http://localhost:8080" {
			ctxName = "custom"
		}

		found := false
		for i, ctx := range cfg.Contexts {
			if ctx.Name == ctxName {
				cfg.Contexts[i].Server = server
				cfg.Contexts[i].APIKey = loginAPIKey
				found = true
				break
			}
		}

		if !found {
			cfg.Contexts = append(cfg.Contexts, config.Context{
				Name:   ctxName,
				Server: server,
				APIKey: loginAPIKey,
			})
		}

		cfg.CurrentContext = ctxName

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Logged in to %s as context %s\n", server, ctxName)
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&loginServer, "server", "", "Kranix API server URL (default: http://localhost:8080)")
	loginCmd.Flags().StringVar(&loginAPIKey, "api-key", "", "API key for authentication")
	loginCmd.Flags().BoolVar(&loginOIDC, "oidc", false, "Use OIDC browser-based login")
}
