package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/spf13/cobra"
)

var aiNamespace string
var aiContext string
var aiMode string

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "Inline AI assistant for analysis and fixes",
	Long:  "Launch an interactive AI assistant that analyzes your context, diagnoses issues, and suggests fixes directly in the terminal.",
	Example: `  kranix ai
  kranix ai --analyze my-app
  kranix ai --context "my-app is crashing with OOM errors"
  kranix ai --mode chat`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, creds, err := getCredentials()
		if err != nil {
			return err
		}

		namespace := aiNamespace
		if namespace == "" {
			namespace = cfg.Defaults.Namespace
		}

		cli := client.New(creds.Server, creds.APIKey)

		if aiContext != "" {
			// One-shot analysis with context
			return analyzeWithContext(cli, namespace, aiContext)
		}

		if len(args) > 0 && args[0] == "--analyze" {
			// Analyze specific workload
			if len(args) < 2 {
				return fmt.Errorf("workload name required for --analyze")
			}
			return analyzeWorkload(cli, args[1], namespace)
		}

		// Interactive chat mode
		return runInteractiveAI(cli, namespace)
	},
}

func init() {
	aiCmd.Flags().StringVar(&aiNamespace, "namespace", "", "Target namespace")
	aiCmd.Flags().StringVar(&aiContext, "context", "", "Provide context for one-shot analysis")
	aiCmd.Flags().StringVar(&aiMode, "mode", "chat", "Mode: chat, analyze, fix")
}

func analyzeWithContext(cli *client.Client, namespace, prompt string) error {
	response, err := cli.AskAI(context.Background(), namespace, prompt)
	if err != nil {
		return fmt.Errorf("failed to get AI response: %w", err)
	}

	fmt.Println("AI Analysis:")
	fmt.Println(response.Response)
	if response.SuggestedAction != "" {
		fmt.Printf("\nSuggested Action: %s\n", response.SuggestedAction)
	}

	return nil
}

func analyzeWorkload(cli *client.Client, workloadName, namespace string) error {
	// Get current status first
	status, err := cli.GetStatus(context.Background(), workloadName, namespace)
	if err != nil {
		return fmt.Errorf("failed to get workload status: %w", err)
	}

	prompt := fmt.Sprintf("Analyze workload %s in namespace %s. Current state: %s, Image: %s. Identify any issues and suggest fixes.",
		workloadName, namespace, status.State, status.Image)

	response, err := cli.AskAI(context.Background(), namespace, prompt)
	if err != nil {
		return fmt.Errorf("failed to get AI analysis: %w", err)
	}

	fmt.Println("AI Analysis:")
	fmt.Println(response.Response)
	if response.SuggestedAction != "" {
		fmt.Printf("\nSuggested Action: %s\n", response.SuggestedAction)
	}
	if response.CodeSnippet != "" {
		fmt.Printf("\nCode Snippet:\n%s\n", response.CodeSnippet)
	}

	return nil
}

func runInteractiveAI(cli *client.Client, namespace string) error {
	fmt.Println("Kranix AI Assistant")
	fmt.Println("Type your questions or issues. Type 'quit' or 'exit' to leave.")
	fmt.Println(strings.Repeat("-", 50))

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if strings.ToLower(input) == "quit" || strings.ToLower(input) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		response, err := cli.AskAI(context.Background(), namespace, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\nAI: %s\n", response.Response)
		if response.SuggestedAction != "" {
			fmt.Printf("\n💡 Suggested: %s\n", response.SuggestedAction)
		}
		if response.CodeSnippet != "" {
			fmt.Printf("\n📝 Code:\n%s\n", response.CodeSnippet)
		}
	}

	return nil
}
