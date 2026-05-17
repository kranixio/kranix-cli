package commands

import (
	"fmt"
	"strings"

	"github.com/kranix-io/kranix-cli/internal/config"
	"github.com/kranix-io/kranix-cli/internal/output"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage server contexts",
	Long:  "List, switch, and set default server contexts.",
}

var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved server contexts",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(cfg.Contexts) == 0 {
			fmt.Println("No contexts configured.")
			return nil
		}

		fmt.Printf("%-20s %-30s\n", "NAME", "SERVER")
		fmt.Println(strings.Repeat("-", 50))
		for _, ctx := range cfg.Contexts {
			marker := ""
			if ctx.Name == cfg.CurrentContext {
				marker = "*"
			}
			fmt.Printf("%s%-20s %-30s\n", marker, ctx.Name, ctx.Server)
		}

		return nil
	},
}

var contextUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch active context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		name := args[0]
		if config.GetContext(cfg, name) == nil {
			return fmt.Errorf("context %q not found", name)
		}

		cfg.CurrentContext = name
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Switched to context %q", name))
		return nil
	},
}

var contextSetDefaultCmd = &cobra.Command{
	Use:   "set-default <name>",
	Short: "Set default context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		name := args[0]
		if config.GetContext(cfg, name) == nil {
			return fmt.Errorf("context %q not found", name)
		}

		cfg.CurrentContext = name
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Default context set to %q", name))
		return nil
	},
}

var contextProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage environment profiles (dev, staging, prod)",
	Long:  "Quickly switch between pre-configured environment profiles like dev, staging, and prod.",
}

var contextProfileSetCmd = &cobra.Command{
	Use:   "set <profile> <server> [--api-key <key>]",
	Short: "Set or update a profile",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		server := args[1]
		apiKey := contextProfileAPIKey

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Update or add the context
		found := false
		for i, ctx := range cfg.Contexts {
			if ctx.Name == profileName {
				cfg.Contexts[i].Server = server
				if apiKey != "" {
					cfg.Contexts[i].APIKey = apiKey
				}
				found = true
				break
			}
		}

		if !found {
			newCtx := config.Context{
				Name:   profileName,
				Server: server,
				APIKey: apiKey,
			}
			cfg.Contexts = append(cfg.Contexts, newCtx)
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Profile %q %s", profileName, map[bool]string{true: "updated", false: "created"}[found]))
		return nil
	},
}

var contextProfileSwitchCmd = &cobra.Command{
	Use:   "switch <profile>",
	Short: "Switch to a profile (dev, staging, prod)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if config.GetContext(cfg, profileName) == nil {
			return fmt.Errorf("profile %q not found. Use 'kranix context profile set' to create it first", profileName)
		}

		cfg.CurrentContext = profileName
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Switched to profile: %s", profileName))
		fmt.Printf("Server: %s\n", config.GetContext(cfg, profileName).Server)
		return nil
	},
}

var contextProfileAPIKey string

func init() {
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextUseCmd)
	contextCmd.AddCommand(contextSetDefaultCmd)
	contextCmd.AddCommand(contextProfileCmd)
	contextProfileCmd.AddCommand(contextProfileSetCmd)
	contextProfileCmd.AddCommand(contextProfileSwitchCmd)

	contextProfileSetCmd.Flags().StringVar(&contextProfileAPIKey, "api-key", "", "API key for the profile")
}
