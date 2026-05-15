package commands

import (
	"context"
	"fmt"

	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/kranix-io/kranix-cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteNamespace string
var deleteConfirm bool

var deleteCmd = &cobra.Command{
	Use:   "delete <workload-name>",
	Short: "Delete a workload",
	Long:  "Delete a container workload from the platform.",
	Example: `  kranix delete my-app
  kranix delete my-app --namespace staging --confirm`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, creds, err := getCredentials()
		if err != nil {
			return err
		}

		name := args[0]
		namespace := deleteNamespace
		if namespace == "" {
			namespace = cfg.Defaults.Namespace
		}

		if !deleteConfirm {
			fmt.Printf("Are you sure you want to delete workload %s? [y/N]: ", name)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Deletion cancelled")
				return nil
			}
		}

		cli := client.New(creds.Server, creds.APIKey)

		if err := cli.Delete(context.Background(), name, namespace); err != nil {
			return fmt.Errorf("failed to delete workload: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Workload %s deleted successfully", name))
		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVar(&deleteNamespace, "namespace", "", "Target namespace")
	deleteCmd.Flags().BoolVar(&deleteConfirm, "confirm", false, "Skip interactive confirmation prompt")
}
