package commands

import (
	"context"
	"fmt"

	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/kranix-io/kranix-cli/internal/output"
	"github.com/spf13/cobra"
)

var costNamespace string
var costOutput string
var costDuration string
var costWorkload string

var costCmd = &cobra.Command{
	Use:   "cost [workload-name]",
	Short: "Show estimated spend breakdown per workload",
	Long:  "Display cost analysis and estimated spend breakdown for workloads, including compute, storage, and network costs.",
	Example: `  kranix cost
  kranix cost my-app
  kranix cost --namespace staging --duration 30d
  kranix cost --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, creds, err := getCredentials()
		if err != nil {
			return err
		}

		namespace := costNamespace
		if namespace == "" {
			namespace = cfg.Defaults.Namespace
		}

		cli := client.New(creds.Server, creds.APIKey)

		format := output.FormatTable
		if costOutput == "json" {
			format = output.FormatJSON
		} else if costOutput == "yaml" {
			format = output.FormatYAML
		}

		duration := costDuration
		if duration == "" {
			duration = "7d" // Default to 7 days
		}

		if len(args) > 0 {
			// Get cost for specific workload
			workloadName := args[0]
			return getWorkloadCost(cli, workloadName, namespace, duration, format)
		}

		// Get cost summary for all workloads
		return getCostSummary(cli, namespace, duration, format)
	},
}

func init() {
	costCmd.Flags().StringVar(&costNamespace, "namespace", "", "Filter by namespace")
	costCmd.Flags().StringVarP(&costOutput, "output", "o", "table", "Output format: table, json, yaml")
	costCmd.Flags().StringVarP(&costDuration, "duration", "d", "7d", "Time duration (e.g., 1h, 1d, 7d, 30d)")
	costCmd.Flags().StringVar(&costWorkload, "workload", "", "Specific workload name")
}

func getWorkloadCost(cli *client.Client, workloadName, namespace, duration string, format output.Format) error {
	costData, err := cli.GetWorkloadCost(context.Background(), workloadName, namespace, duration)
	if err != nil {
		return fmt.Errorf("failed to get workload cost: %w", err)
	}

	output.PrintSuccess(fmt.Sprintf("Cost breakdown for %s (last %s)", workloadName, duration))
	fmt.Printf("  Total Estimated Cost: $%.2f\n", costData.TotalCost)
	fmt.Printf("  Compute Cost: $%.2f\n", costData.ComputeCost)
	fmt.Printf("  Storage Cost: $%.2f\n", costData.StorageCost)
	fmt.Printf("  Network Cost: $%.2f\n", costData.NetworkCost)

	if len(costData.Breakdown) > 0 {
		fmt.Println("\n  Resource Breakdown:")
		for _, item := range costData.Breakdown {
			fmt.Printf("    - %s: $%.2f (%s)\n", item.Resource, item.Cost, item.Usage)
		}
	}

	return output.Print(format, costData)
}

func getCostSummary(cli *client.Client, namespace, duration string, format output.Format) error {
	summary, err := cli.GetCostSummary(context.Background(), namespace, duration)
	if err != nil {
		return fmt.Errorf("failed to get cost summary: %w", err)
	}

	output.PrintSuccess(fmt.Sprintf("Cost summary for namespace %s (last %s)", namespace, duration))
	fmt.Printf("  Total Estimated Cost: $%.2f\n", summary.TotalCost)
	fmt.Printf("  Workloads: %d\n", summary.WorkloadCount)
	fmt.Printf("  Average per Workload: $%.2f\n", summary.AverageCost)

	if len(summary.TopCostWorkloads) > 0 {
		fmt.Println("\n  Top Cost Workloads:")
		for i, w := range summary.TopCostWorkloads {
			fmt.Printf("    %d. %s: $%.2f\n", i+1, w.WorkloadName, w.TotalCost)
		}
	}

	return output.Print(format, summary)
}
