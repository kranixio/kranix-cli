package completion

import (
	"github.com/spf13/cobra"
)

// ValidArgsFunc returns a function that provides completion for context names
func ValidContextsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO: Load config and return context names for completion
	return []string{}, cobra.ShellCompDirectiveNoFileComp
}

// ValidNamespacesFunc returns a function that provides completion for namespace names
func ValidNamespacesFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO: Fetch namespaces from API for completion
	return []string{}, cobra.ShellCompDirectiveNoFileComp
}

// ValidWorkloadsFunc returns a function that provides completion for workload names
func ValidWorkloadsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO: Fetch workloads from API for completion
	return []string{}, cobra.ShellCompDirectiveNoFileComp
}
