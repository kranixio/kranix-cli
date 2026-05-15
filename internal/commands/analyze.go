package commands

import (
	"context"
	"fmt"

	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/kranix-io/kranix-cli/internal/output"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:     "analyze <workload-name>",
	Short:   "Analyze a failing workload",
	Long:    "Run AI-powered failure analysis via kranix-api. Returns crash reason, probable fix, and resource bottleneck detection.",
	Example: `  kranix analyze my-app`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, creds, err := getCredentials()
		if err != nil {
			return err
		}

		name := args[0]
		namespace := cfg.Defaults.Namespace

		cli := client.New(creds.Server, creds.APIKey)

		// TODO: Implement analyze endpoint in kranix-api
		// For now, just get status
		status, err := cli.GetStatus(context.Background(), name, namespace)
		if err != nil {
			return fmt.Errorf("failed to analyze workload: %w", err)
		}

		fmt.Println("Analysis results:")
		fmt.Printf("  Workload: %s\n", name)
		fmt.Printf("  State: %s\n", status.State)
		fmt.Println("\n  AI-powered analysis is not yet implemented.")
		fmt.Println("  This will connect to kranix-api's analysis endpoint once available.")

		return output.Print(output.FormatTable, status)
	},
}
