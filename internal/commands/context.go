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

func init() {
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextUseCmd)
	contextCmd.AddCommand(contextSetDefaultCmd)
}
