package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/spf13/cobra"
)

var dashboardNamespace string
var dashboardRefresh int

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch TUI live dashboard (k9s-style)",
	Long:  "Launch a terminal UI dashboard for real-time cluster state monitoring with k9s-style navigation.",
	Example: `  kranix dashboard
  kranix dashboard --namespace staging
  kranix dashboard --refresh 2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, creds, err := getCredentials()
		if err != nil {
			return err
		}

		namespace := dashboardNamespace
		if namespace == "" {
			namespace = cfg.Defaults.Namespace
		}

		refreshInterval := time.Duration(dashboardRefresh) * time.Second
		if refreshInterval == 0 {
			refreshInterval = 5 * time.Second
		}

		cli := client.New(creds.Server, creds.APIKey)

		model := newDashboardModel(cli, namespace, refreshInterval)
		program := tea.NewProgram(model, tea.WithAltScreen())

		if _, err := program.Run(); err != nil {
			return fmt.Errorf("failed to run dashboard: %w", err)
		}

		return nil
	},
}

func init() {
	dashboardCmd.Flags().StringVar(&dashboardNamespace, "namespace", "", "Filter by namespace")
	dashboardCmd.Flags().IntVar(&dashboardRefresh, "refresh", 5, "Refresh interval in seconds")
}

// Dashboard model for tea
type dashboardModel struct {
	cli             *client.Client
	namespace       string
	refreshInterval time.Duration
	workloads       []*client.WorkloadStatus
	loading         bool
	err             error
	selectedIndex   int
	quitting        bool
}

type workloadsLoadedMsg struct {
	workloads []*client.WorkloadStatus
	err       error
}

type tickMsg time.Time

func newDashboardModel(cli *client.Client, namespace string, refreshInterval time.Duration) dashboardModel {
	return dashboardModel{
		cli:             cli,
		namespace:       namespace,
		refreshInterval: refreshInterval,
		loading:         true,
		selectedIndex:   0,
		quitting:        false,
	}
}

func (m dashboardModel) Init() tea.Cmd {
	return m.loadWorkloads()
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case workloadsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.workloads = msg.workloads
		return m, m.waitForNextTick()

	case tickMsg:
		m.loading = true
		return m, m.loadWorkloads()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case "down", "j":
			if m.selectedIndex < len(m.workloads)-1 {
				m.selectedIndex++
			}
		case "enter":
			if len(m.workloads) > 0 {
				// Show details for selected workload
				selected := m.workloads[m.selectedIndex]
				fmt.Printf("\nSelected: %s (State: %s, Image: %s)\n", selected.Name, selected.State, selected.Image)
			}
		}
	}

	return m, nil
}

func (m dashboardModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var title string
	if m.loading {
		title = loadingStyle.Render("Loading workloads...")
	} else {
		title = titleStyle.Render(fmt.Sprintf("Kranix Dashboard - Namespace: %s", m.namespace))
	}

	header := title + "\n\n"

	if m.loading {
		return header + loadingStyle.Render("Fetching cluster state...")
	}

	if len(m.workloads) == 0 {
		return header + "No workloads found in namespace.\n"
	}

	// Build table
	table := ""
	for i, w := range m.workloads {
		prefix := " "
		if i == m.selectedIndex {
			prefix = ">"
		}
		row := fmt.Sprintf("%s %-20s %-15s %s\n", prefix, w.Name, w.State, w.Image)
		if i == m.selectedIndex {
			table += selectedStyle.Render(row)
		} else {
			table += row
		}
	}

	footer := "\n" + helpStyle.Render("↑/k: up | ↓/j: down | enter: details | q: quit")

	return header + table + footer
}

func (m dashboardModel) loadWorkloads() tea.Cmd {
	return func() tea.Msg {
		workloads, err := m.cli.ListWorkloads(context.Background(), m.namespace)
		return workloadsLoadedMsg{workloads: workloads, err: err}
	}
}

func (m dashboardModel) waitForNextTick() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Styles
var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	loadingStyle  = lipgloss.NewStyle().Faint(true)
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F5F"))
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#7D56F4")).Foreground(lipgloss.Color("#FFFFFF"))
	helpStyle     = lipgloss.NewStyle().Faint(true)
)



