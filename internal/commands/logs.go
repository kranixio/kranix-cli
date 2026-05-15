package commands

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/kranix-io/kranix-cli/internal/output"
	"github.com/spf13/cobra"
)

var logsFollow bool
var logsTail int
var logsSince string
var logsPod string
var logsJSON bool

var logsCmd = &cobra.Command{
	Use:   "logs <workload-name>",
	Short: "Stream workload logs",
	Long:  "Stream logs from a workload in real-time.",
	Example: `  kranix logs my-app --follow
  kranix logs my-app --tail 50
  kranix logs my-app --since 10m`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, creds, err := getCredentials()
		if err != nil {
			return err
		}

		name := args[0]
		namespace := cfg.Defaults.Namespace

		cli := client.New(creds.Server, creds.APIKey)
		opts := &client.LogOptions{
			TailLines: logsTail,
			Follow:    logsFollow,
			Since:     logsSince,
		}

		logStream, err := cli.StreamLogs(context.Background(), name, namespace, opts)
		if err != nil {
			return fmt.Errorf("failed to stream logs: %w", err)
		}
		defer logStream.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle Ctrl+C
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()

		scanner := bufio.NewScanner(logStream)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return nil
			default:
				if logsJSON {
					output.PrintLogLine(scanner.Text())
				} else {
					fmt.Println(scanner.Text())
				}
			}
		}

		if err := scanner.Err(); err != nil && err != io.EOF {
			return fmt.Errorf("error reading logs: %w", err)
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Stream logs in real time")
	logsCmd.Flags().IntVar(&logsTail, "tail", 100, "Number of lines to show from end")
	logsCmd.Flags().StringVar(&logsSince, "since", "", "Show logs since a duration (e.g. 10m, 1h)")
	logsCmd.Flags().StringVar(&logsPod, "pod", "", "Target a specific pod")
	logsCmd.Flags().BoolVar(&logsJSON, "json", false, "Output as JSON")
}
