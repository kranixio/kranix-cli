package commands

import (
	"context"
	"fmt"

	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/kranix-io/kranix-cli/internal/output"
	"github.com/spf13/cobra"
)

var namespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "Manage namespaces",
	Long:  "Create, list, and delete namespaces.",
}

var namespaceCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a namespace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, creds, err := getCredentials()
		if err != nil {
			return err
		}

		name := args[0]
		cli := client.New(creds.Server, creds.APIKey)

		if err := cli.CreateNamespace(context.Background(), name); err != nil {
			return fmt.Errorf("failed to create namespace: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Namespace %s created successfully", name))
		return nil
	},
}

var namespaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all namespaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, creds, err := getCredentials()
		if err != nil {
			return err
		}

		cli := client.New(creds.Server, creds.APIKey)

		namespaces, err := cli.ListNamespaces(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list namespaces: %w", err)
		}

		return output.Print(output.FormatTable, namespaces)
	},
}

var namespaceDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a namespace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, creds, err := getCredentials()
		if err != nil {
			return err
		}

		name := args[0]
		cli := client.New(creds.Server, creds.APIKey)

		if err := cli.DeleteNamespace(context.Background(), name); err != nil {
			return fmt.Errorf("failed to delete namespace: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Namespace %s deleted successfully", name))
		return nil
	},
}

func init() {
	namespaceCmd.AddCommand(namespaceCreateCmd)
	namespaceCmd.AddCommand(namespaceListCmd)
	namespaceCmd.AddCommand(namespaceDeleteCmd)
}
