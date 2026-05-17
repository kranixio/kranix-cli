package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/kranix-io/kranix-cli/internal/config"
	"github.com/spf13/cobra"
)

var pfNamespace string
var pfLocalPort int
var pfRemotePort int
var pfPod string

var portForwardCmd = &cobra.Command{
	Use:   "port-forward <workload-name>",
	Short: "Forward local port to workload pod",
	Long:  "Shorthand for forwarding a local port to any workload pod. Automatically selects a pod and forwards the specified port.",
	Example: `  kranix port-forward my-app --local 8080 --remote 80
  kranix port-forward my-app --local 3000
  kranix port-forward my-app --namespace staging --pod my-app-abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, creds, err := getCredentials()
		if err != nil {
			return err
		}

		namespace := pfNamespace
		if namespace == "" {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			namespace = cfg.Defaults.Namespace
		}

		cli := client.New(creds.Server, creds.APIKey)
		workloadName := args[0]

		// Get workload pods
		pods, err := cli.ListPods(context.Background(), workloadName, namespace)
		if err != nil {
			return fmt.Errorf("failed to list pods: %w", err)
		}

		if len(pods) == 0 {
			return fmt.Errorf("no pods found for workload %s", workloadName)
		}

		// Select pod
		selectedPod := pfPod
		if selectedPod == "" {
			selectedPod = pods[0].Name
		} else {
			// Verify pod exists
			found := false
			for _, p := range pods {
				if p.Name == selectedPod {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("pod %s not found for workload %s", selectedPod, workloadName)
			}
		}

		// Set default ports
		localPort := pfLocalPort
		if localPort == 0 {
			localPort = 8080
		}

		remotePort := pfRemotePort
		if remotePort == 0 {
			remotePort = 80
		}

		fmt.Printf("Forwarding from localhost:%d to pod %s:%d\n", localPort, selectedPod, remotePort)
		fmt.Printf("Workload: %s, Namespace: %s\n", workloadName, namespace)
		fmt.Println("Press Ctrl+C to stop forwarding")

		// Execute port-forward using kubectl (simplified approach)
		// In production, this would use kranix-api to get kubeconfig and exec kubectl
		return executePortForward(selectedPod, namespace, localPort, remotePort)
	},
}

func init() {
	portForwardCmd.Flags().StringVar(&pfNamespace, "namespace", "", "Target namespace")
	portForwardCmd.Flags().IntVarP(&pfLocalPort, "local", "l", 0, "Local port (default: 8080)")
	portForwardCmd.Flags().IntVarP(&pfRemotePort, "remote", "r", 0, "Remote pod port (default: 80)")
	portForwardCmd.Flags().StringVarP(&pfPod, "pod", "p", "", "Specific pod name (auto-selects if not specified)")
}

func executePortForward(podName, namespace string, localPort, remotePort int) error {
	// Check if kubectl is available
	if _, err := exec.LookPath("kubectl"); err != nil {
		return fmt.Errorf("kubectl not found. Please install kubectl to use port-forwarding")
	}

	// Build kubectl port-forward command
	args := []string{
		"port-forward",
		fmt.Sprintf("pod/%s", podName),
		fmt.Sprintf("%d:%d", localPort, remotePort),
		"-n", namespace,
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
