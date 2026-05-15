package commands

import (
	"context"
	"fmt"

	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/kranix-io/kranix-cli/internal/output"
	"github.com/spf13/cobra"
)

var restartNamespace string

var restartCmd = &cobra.Command{
	Use:   "restart <workload-name>",
	Short: "Restart a workload",
	Long:  "Restart a container workload.",
	Example: `  kranix restart my-app
  kranix restart my-app --namespace staging`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, creds, err := getCredentials()
		if err != nil {
			return err
		}

		name := args[0]
		namespace := restartNamespace
		if namespace == "" {
			namespace = cfg.Defaults.Namespace
		}

		cli := client.New(creds.Server, creds.APIKey)

		if err := cli.Restart(context.Background(), name, namespace); err != nil {
			return fmt.Errorf("failed to restart workload: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Workload %s restarted successfully", name))
		return nil
	},
}

func init() {
	restartCmd.Flags().StringVar(&restartNamespace, "namespace", "", "Target namespace")
}
