package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/kranix-io/kranix-cli/internal/output"
	"github.com/spf13/cobra"
)

var diffNamespace string
var diffOutput string
var diffFile string

var diffCmd = &cobra.Command{
	Use:   "diff [workload-name]",
	Short: "Show exactly what would change before applying",
	Long:  "Preview changes to a workload before applying them. This uses dry-run mode to show the diff between current and proposed state.",
	Example: `  kranix diff my-app --image nginx:1.25
  kranix diff --file manifest.yaml
  kranix diff my-app --replicas 3 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, creds, err := getCredentials()
		if err != nil {
			return err
		}

		namespace := diffNamespace
		if namespace == "" {
			namespace = cfg.Defaults.Namespace
		}

		cli := client.New(creds.Server, creds.APIKey)

		format := output.FormatTable
		if diffOutput == "json" {
			format = output.FormatJSON
		} else if diffOutput == "yaml" {
			format = output.FormatYAML
		}

		if diffFile != "" {
			// Diff from file
			return diffFromFile(cli, namespace, diffFile, format)
		}

		if len(args) > 0 {
			// Diff specific workload with flags
			workloadName := args[0]
			return diffWorkload(cli, workloadName, namespace, format)
		}

		return fmt.Errorf("either provide a workload name or use --file")
	},
}

func init() {
	diffCmd.Flags().StringVar(&diffNamespace, "namespace", "", "Target namespace")
	diffCmd.Flags().StringVarP(&diffOutput, "output", "o", "table", "Output format: table, json, yaml")
	diffCmd.Flags().StringVarP(&diffFile, "file", "f", "", "Diff from manifest file")
	diffCmd.Flags().StringVar(&deployImage, "image", "", "New container image")
	diffCmd.Flags().IntVar(&deployReplicas, "replicas", 0, "New replica count")
	diffCmd.Flags().StringSliceVar(&deployEnv, "env", []string{}, "Environment variables (KEY=VALUE)")
}

func diffFromFile(cli *client.Client, namespace, filePath string, format output.Format) error {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the file (simplified - in production, use proper YAML/JSON parsing)
	fmt.Printf("Diffing from file: %s\n", filePath)
	fmt.Printf("File content preview:\n%s\n", string(data[:min(len(data), 500)]))

	// Call dry-run preview endpoint
	preview, err := cli.GetDryRunPreview(context.Background(), namespace, string(data))
	if err != nil {
		return fmt.Errorf("failed to get dry-run preview: %w", err)
	}

	output.PrintSuccess("Dry-run preview:")
	return output.Print(format, preview)
}

func diffWorkload(cli *client.Client, workloadName, namespace string, format output.Format) error {
	// Get current state
	current, err := cli.GetStatus(context.Background(), workloadName, namespace)
	if err != nil {
		return fmt.Errorf("failed to get current workload status: %w", err)
	}

	// Build proposed spec from flags
	proposedSpec := &client.WorkloadSpec{
		Name:      workloadName,
		Namespace: namespace,
		Image:     current.Image,
		Replicas:  1, // Default
	}

	if deployImage != "" {
		proposedSpec.Image = deployImage
	}
	if deployReplicas > 0 {
		proposedSpec.Replicas = deployReplicas
	}

	envMap := make(map[string]string)
	for _, e := range deployEnv {
		parts := splitEnv(e)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	proposedSpec.Env = envMap

	// Get diff from API
	diff, err := cli.GetDiff(context.Background(), workloadName, namespace, proposedSpec)
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}

	output.PrintSuccess(fmt.Sprintf("Diff for workload: %s", workloadName))
	fmt.Printf("Current state:\n")
	fmt.Printf("  Name: %s\n", current.Name)
	fmt.Printf("  Image: %s\n", current.Image)
	fmt.Printf("  State: %s\n", current.State)

	fmt.Printf("\nProposed changes:\n")
	return output.Print(format, diff)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
