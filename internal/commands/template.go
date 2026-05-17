package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/spf13/cobra"
)

var templateList bool
var templateName string
var templateOutput string
var templateVars []string

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Bootstrap a new workload from community templates",
	Long:  "Create workloads from predefined community templates. List available templates or apply a template with custom variables.",
	Example: `  kranix template --list
  kranix template --name nginx --output my-app.yaml
  kranix template --name nodejs --var PORT=8080 --var NODE_ENV=production`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, creds, err := getCredentials()
		if err != nil {
			return err
		}

		cli := client.New(creds.Server, creds.APIKey)

		if templateList {
			return listTemplates(cli)
		}

		if templateName == "" {
			return fmt.Errorf("--name is required when not using --list")
		}

		return applyTemplate(cli, templateName, templateOutput, templateVars)
	},
}

func init() {
	templateCmd.Flags().BoolVar(&templateList, "list", false, "List available templates")
	templateCmd.Flags().StringVar(&templateName, "name", "", "Template name to apply")
	templateCmd.Flags().StringVarP(&templateOutput, "output", "o", "", "Output file path (default: stdout)")
	templateCmd.Flags().StringSliceVar(&templateVars, "var", []string{}, "Template variables (KEY=VALUE)")
}

func listTemplates(cli *client.Client) error {
	templates, err := cli.ListTemplates(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	fmt.Println("Available Templates:")
	fmt.Println("-------------------")
	for _, t := range templates {
		fmt.Printf("  %-20s %s\n", t.Name, t.Description)
		if t.Category != "" {
			fmt.Printf("    Category: %s\n", t.Category)
		}
	}

	return nil
}

func applyTemplate(cli *client.Client, templateName, outputPath string, vars []string) error {
	// Parse variables
	varMap := make(map[string]string)
	for _, v := range vars {
		parts := splitEnv(v)
		if len(parts) == 2 {
			varMap[parts[0]] = parts[1]
		}
	}

	// Get template from API
	template, err := cli.GetTemplate(context.Background(), templateName, varMap)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// Output the template
	if outputPath != "" {
		// Ensure directory exists
		dir := filepath.Dir(outputPath)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		}

		if err := os.WriteFile(outputPath, []byte(template.Content), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Template applied successfully to: %s\n", outputPath)
	} else {
		fmt.Println(template.Content)
	}

	return nil
}
