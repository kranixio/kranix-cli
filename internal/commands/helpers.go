package commands

import (
	"fmt"

	"github.com/kranix-io/kranix-cli/internal/auth"
	"github.com/kranix-io/kranix-cli/internal/config"
)

func getCredentials() (*config.Config, *auth.Credentials, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Try environment variables first
	if creds, err := auth.GetCredentialsFromEnv(); err == nil {
		return cfg, creds, nil
	}

	// Fall back to config file context
	ctx := config.GetCurrentContext(cfg)
	if ctx == nil {
		return nil, nil, fmt.Errorf("no context configured. Run 'kranix login' first")
	}

	creds := &auth.Credentials{
		Server: ctx.Server,
		APIKey: ctx.APIKey,
	}

	return cfg, creds, nil
}

func splitEnv(env string) []string {
	// Split KEY=VALUE into [KEY, VALUE]
	parts := make([]string, 0, 2)
	for i, r := range env {
		if r == '=' {
			parts = append(parts, env[:i], env[i+1:])
			return parts
		}
	}
	return []string{env}
}

func getNamespace(cfg *config.Config, flag string) string {
	if flag != "" {
		return flag
	}
	if cfg.Defaults.Namespace != "" {
		return cfg.Defaults.Namespace
	}
	return "default"
}
