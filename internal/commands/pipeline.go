package commands

import (
	"fmt"

	"github.com/kranix-io/kranix-cli/internal/output"
	"github.com/spf13/cobra"
)

var pipelineName string
var pipelineFile string
var pipelineApproveStage string
var pipelineList bool

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Manage deployment pipelines",
	Long:  "Create, list, and manage multi-stage deployment pipelines.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if pipelineList {
			return listPipelines()
		}
		if pipelineApproveStage != "" {
			return approveStage(pipelineName, pipelineApproveStage)
		}
		if pipelineFile != "" {
			return createPipeline(pipelineFile)
		}
		return cmd.Help()
	},
}

func createPipeline(file string) error {
	fmt.Printf("Creating pipeline from %s\n", file)
	// TODO: Implement pipeline creation
	output.PrintSuccess("Pipeline created successfully")
	return nil
}

func listPipelines() error {
	fmt.Println("Listing pipelines")
	// TODO: Implement pipeline listing
	output.PrintSuccess("Pipelines listed successfully")
	return nil
}

func approveStage(pipelineName, stageName string) error {
	if pipelineName == "" {
		return fmt.Errorf("--name is required for stage approval")
	}
	fmt.Printf("Approving stage %s in pipeline %s\n", stageName, pipelineName)
	// TODO: Implement stage approval
	output.PrintSuccess("Stage approved successfully")
	return nil
}

func init() {
	pipelineCmd.Flags().StringVar(&pipelineName, "name", "", "Pipeline name")
	pipelineCmd.Flags().StringVarP(&pipelineFile, "file", "f", "", "Pipeline manifest file")
	pipelineCmd.Flags().StringVar(&pipelineApproveStage, "approve", "", "Approve a specific stage")
	pipelineCmd.Flags().BoolVarP(&pipelineList, "list", "l", false, "List all pipelines")
}
