package commands

import (
	"context"
	"fmt"

	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/kranix-io/kranix-cli/internal/output"
	"github.com/spf13/cobra"
)

var deployName string
var deployImage string
var deployNamespace string
var deployReplicas int
var deployEnv []string
var deployPort int
var deployCPU string
var deployMemory string
var deployWait bool
var deployStrategy string
var deployCanaryReplicas int
var deployCanaryPercentage int
var deployCanaryAutoPromote bool
var deployFederationEnabled bool
var deployFederationClusters []string
var deployAutoRollback bool
var deployRollbackThreshold float64

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a workload",
	Long:  "Deploy a container workload to the Kranix platform.",
	Example: `  kranix deploy --name my-app --image nginx:latest --namespace staging
  kranix deploy --name my-app --image nginx:latest --env DB_HOST=localhost --replicas 3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if deployName == "" {
			return fmt.Errorf("--name is required")
		}
		if deployImage == "" {
			return fmt.Errorf("--image is required")
		}

		cfg, creds, err := getCredentials()
		if err != nil {
			return err
		}

		namespace := deployNamespace
		if namespace == "" {
			namespace = cfg.Defaults.Namespace
		}

		envMap := make(map[string]string)
		for _, e := range deployEnv {
			parts := splitEnv(e)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}

		cli := client.New(creds.Server, creds.APIKey)
		spec := &client.WorkloadSpec{
			Name:      deployName,
			Image:     deployImage,
			Namespace: namespace,
			Replicas:  deployReplicas,
			Env:       envMap,
			Port:      deployPort,
			CPU:       deployCPU,
			Memory:    deployMemory,
		}

		status, err := cli.Deploy(context.Background(), spec)
		if err != nil {
			return fmt.Errorf("failed to deploy workload: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("Workload %s deployed successfully", deployName))
		return output.Print(output.FormatTable, status)
	},
}

func init() {
	deployCmd.Flags().StringVar(&deployName, "name", "", "Workload name (required)")
	deployCmd.Flags().StringVar(&deployImage, "image", "", "Container image (required)")
	deployCmd.Flags().StringVar(&deployNamespace, "namespace", "", "Target namespace (default: default)")
	deployCmd.Flags().IntVar(&deployReplicas, "replicas", 1, "Number of replicas")
	deployCmd.Flags().StringSliceVar(&deployEnv, "env", []string{}, "Environment variables (KEY=VALUE, repeatable)")
	deployCmd.Flags().IntVar(&deployPort, "port", 0, "Exposed port")
	deployCmd.Flags().StringVar(&deployCPU, "cpu", "", "CPU limit (e.g. 500m)")
	deployCmd.Flags().StringVar(&deployMemory, "memory", "", "Memory limit (e.g. 256Mi)")
	deployCmd.Flags().BoolVar(&deployWait, "wait", false, "Wait for workload to become ready")
	// Progressive delivery flags
	deployCmd.Flags().StringVar(&deployStrategy, "strategy", "rolling", "Deployment strategy: rolling, canary, blue-green")
	deployCmd.Flags().IntVar(&deployCanaryReplicas, "canary-replicas", 1, "Number of canary replicas")
	deployCmd.Flags().IntVar(&deployCanaryPercentage, "canary-percentage", 10, "Percentage of traffic to canary (0-100)")
	deployCmd.Flags().BoolVar(&deployCanaryAutoPromote, "canary-auto-promote", false, "Automatically promote canary on success")
	// Federation flags
	// Auto-rollback flags
	deployCmd.Flags().BoolVar(&deployAutoRollback, "auto-rollback", false, "Enable automatic rollback on failure")
	deployCmd.Flags().Float64Var(&deployRollbackThreshold, "rollback-threshold", 0.05, "Error rate threshold for auto-rollback (0-1)")
	deployCmd.Flags().BoolVar(&deployFederationEnabled, "federation", false, "Enable multi-cluster federation")
	deployCmd.Flags().StringSliceVar(&deployFederationClusters, "federation-clusters", []string{}, "Target clusters for federation")
}
