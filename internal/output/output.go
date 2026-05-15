package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kranix-io/kranix-cli/internal/client"
	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

func Print(format Format, data interface{}) error {
	switch format {
	case FormatJSON:
		return printJSON(data)
	case FormatYAML:
		return printYAML(data)
	case FormatTable:
		return printTable(data)
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
}

func printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func printYAML(data interface{}) error {
	return yaml.NewEncoder(os.Stdout).Encode(data)
}

func printTable(data interface{}) error {
	switch v := data.(type) {
	case []*client.WorkloadStatus:
		return printWorkloadTable(v)
	case []*client.Namespace:
		return printNamespaceTable(v)
	case *client.WorkloadStatus:
		return printWorkloadDetail(v)
	default:
		return printJSON(data)
	}
}

func printWorkloadTable(workloads []*client.WorkloadStatus) error {
	if len(workloads) == 0 {
		fmt.Println("No workloads found.")
		return nil
	}

	fmt.Printf("%-30s %-15s %-30s\n", "NAME", "STATE", "IMAGE")
	fmt.Println(strings.Repeat("-", 75))

	for _, w := range workloads {
		image := w.Image
		if len(image) > 28 {
			image = image[:25] + "..."
		}
		fmt.Printf("%-30s %-15s %-30s\n", w.Name, w.State, image)
	}

	return nil
}

func printNamespaceTable(namespaces []*client.Namespace) error {
	if len(namespaces) == 0 {
		fmt.Println("No namespaces found.")
		return nil
	}

	fmt.Println("NAME")
	fmt.Println(strings.Repeat("-", 20))

	for _, ns := range namespaces {
		fmt.Println(ns.Name)
	}

	return nil
}

func printWorkloadDetail(w *client.WorkloadStatus) error {
	fmt.Printf("Name:   %s\n", w.Name)
	fmt.Printf("ID:     %s\n", w.ID)
	fmt.Printf("State:  %s\n", w.State)
	fmt.Printf("Image:  %s\n", w.Image)
	return nil
}

func PrintLogLine(line string) {
	fmt.Println(line)
}

func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

func PrintSuccess(msg string) {
	fmt.Fprintf(os.Stdout, "✓ %s\n", msg)
}
