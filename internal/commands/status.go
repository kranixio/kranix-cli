package commands

import (
	"context"
	"fmt"

	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/kranix-io/kranix-cli/internal/output"
	"github.com/spf13/cobra"
)

var statusNamespace string
var statusWatch bool
var statusJSON bool

var statusCmd = &cobra.Command{
	Use:   "status [workload-name]",
	Short: "Get workload status",
	Long:  "Get the status of a workload or list all workloads in a namespace.",
	Example: `  kranix status my-app
  kranix status --namespace staging
  kranix status my-app --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, creds, err := getCredentials()
		if err != nil {
			return err
		}

		namespace := statusNamespace
		if namespace == "" {
			namespace = cfg.Defaults.Namespace
		}

		cli := client.New(creds.Server, creds.APIKey)

		format := output.FormatTable
		if statusJSON {
			format = output.FormatJSON
		}

		if len(args) > 0 {
			// Get status of specific workload
			name := args[0]
			status, err := cli.GetStatus(context.Background(), name, namespace)
			if err != nil {
				return fmt.Errorf("failed to get status: %w", err)
			}
			return output.Print(format, status)
		}

		// List all workloads
		workloads, err := cli.ListWorkloads(context.Background(), namespace)
		if err != nil {
			return fmt.Errorf("failed to list workloads: %w", err)
		}
		return output.Print(format, workloads)
	},
}

func init() {
	statusCmd.Flags().StringVar(&statusNamespace, "namespace", "", "Filter by namespace")
	statusCmd.Flags().BoolVar(&statusWatch, "watch", false, "Watch for status changes")
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output as JSON")
}
