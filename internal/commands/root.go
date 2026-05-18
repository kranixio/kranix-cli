package commands

import (
	"os"

	clilib "github.com/kranix-io/kranix-packages/cmd/cli-lib"
	"github.com/spf13/cobra"
)

var globalFlags clilib.GlobalFlags

var rootCmd = &cobra.Command{
	Use:   "kranix",
	Short: "Kranix CLI - deploy, inspect, and debug your infrastructure",
	Long: `kranix-cli is the primary command-line interface for the Kranix platform.
It wraps kranix-api in a fast, developer-friendly shell experience with commands
for deploying workloads, streaming logs, inspecting cluster state, analyzing failures,
and managing namespaces.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := clilib.ValidateFlags(&globalFlags); err != nil {
			clilib.PrintError(err)
			os.Exit(1)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add global flags using the CLI helper library
	clilib.AddGlobalFlags(rootCmd, &globalFlags)

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(namespaceCmd)
	rootCmd.AddCommand(manifestsCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(dashboardCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(aiCmd)
	rootCmd.AddCommand(costCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(offlineCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(portForwardCmd)
	rootCmd.AddCommand(pipelineCmd)
	rootCmd.AddCommand(secretCmd)
	rootCmd.AddCommand(driftCmd)
}
