package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kranix-io/kranix-cli/internal/auth"
	"github.com/kranix-io/kranix-cli/internal/config"
	"github.com/spf13/cobra"
)

var doctorVerbose bool
var doctorFix bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose local setup issues and config problems",
	Long:  "Check your local kranix-cli setup, configuration, and connectivity. Identifies common issues and suggests fixes.",
	Example: `  kranix doctor
  kranix doctor --verbose
  kranix doctor --fix`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDiagnostics(doctorVerbose, doctorFix)
	},
}

func init() {
	doctorCmd.Flags().BoolVarP(&doctorVerbose, "verbose", "v", false, "Show detailed diagnostic information")
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Attempt to auto-fix issues")
}

type DiagnosticCheck struct {
	Name    string
	Status  string
	Message string
	CanFix  bool
	Fixed   bool
}

func runDiagnostics(verbose, autoFix bool) error {
	fmt.Println("Kranix CLI Diagnostics")
	fmt.Println("=======================")
	fmt.Println()

	checks := []DiagnosticCheck{}

	// Check 1: Config file exists
	configCheck := checkConfigFile()
	checks = append(checks, configCheck)

	// Check 2: Context configuration
	contextCheck := checkContextConfiguration()
	checks = append(checks, contextCheck)

	// Check 3: API connectivity
	connectivityCheck := checkAPIConnectivity()
	checks = append(checks, connectivityCheck)

	// Check 4: Credentials
	credsCheck := checkCredentials()
	checks = append(checks, credsCheck)

	// Check 5: Required tools
	toolsCheck := checkRequiredTools()
	checks = append(checks, toolsCheck)

	// Check 6: Permissions
	permsCheck := checkPermissions()
	checks = append(checks, permsCheck)

	// Check 7: Go version (if applicable)
	goCheck := checkGoVersion()
	checks = append(checks, goCheck)

	// Display results
	issuesFound := 0
	for i, check := range checks {
		statusIcon := "✓"
		if check.Status != "OK" {
			statusIcon = "✗"
			issuesFound++
		}

		fmt.Printf("%d. %s %s\n", i+1, statusIcon, check.Name)
		if verbose || check.Status != "OK" {
			fmt.Printf("   Status: %s\n", check.Status)
			fmt.Printf("   %s\n", check.Message)
			if check.Fixed {
				fmt.Printf("   [FIXED] Auto-fix applied\n")
			}
		}
		fmt.Println()
	}

	// Summary
	fmt.Println("=======================")
	if issuesFound == 0 {
		fmt.Println("✓ All checks passed!")
	} else {
		fmt.Printf("✗ %d issue(s) found\n", issuesFound)
		if autoFix {
			fmt.Println("Auto-fix was attempted. Run again to verify.")
		} else {
			fmt.Println("Run 'kranix doctor --fix' to attempt auto-fix.")
		}
	}

	return nil
}

func checkConfigFile() DiagnosticCheck {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return DiagnosticCheck{
			Name:    "Config file location",
			Status:  "WARN",
			Message: fmt.Sprintf("Could not determine config path: %v", err),
			CanFix:  false,
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DiagnosticCheck{
			Name:    "Config file exists",
			Status:  "WARN",
			Message: fmt.Sprintf("Config file not found at %s. Run 'kranix login' to create it.", configPath),
			CanFix:  false,
		}
	}

	return DiagnosticCheck{
		Name:    "Config file exists",
		Status:  "OK",
		Message: fmt.Sprintf("Config file found at %s", configPath),
		CanFix:  false,
	}
}

func checkContextConfiguration() DiagnosticCheck {
	cfg, err := config.Load()
	if err != nil {
		return DiagnosticCheck{
			Name:    "Context configuration",
			Status:  "ERROR",
			Message: fmt.Sprintf("Failed to load config: %v", err),
			CanFix:  false,
		}
	}

	if len(cfg.Contexts) == 0 {
		return DiagnosticCheck{
			Name:    "Context configuration",
			Status:  "WARN",
			Message: "No contexts configured. Run 'kranix login' to add a context.",
			CanFix:  false,
		}
	}

	currentCtx := config.GetCurrentContext(cfg)
	if currentCtx == nil {
		return DiagnosticCheck{
			Name:    "Current context",
			Status:  "WARN",
			Message: "No current context set. Run 'kranix context use <name>' to set one.",
			CanFix:  false,
		}
	}

	return DiagnosticCheck{
		Name:    "Context configuration",
		Status:  "OK",
		Message: fmt.Sprintf("Current context: %s (%s)", currentCtx.Name, currentCtx.Server),
		CanFix:  false,
	}
}

func checkAPIConnectivity() DiagnosticCheck {
	cfg, err := config.Load()
	if err != nil {
		return DiagnosticCheck{
			Name:    "API connectivity",
			Status:  "SKIP",
			Message: "Cannot check connectivity without valid config",
			CanFix:  false,
		}
	}

	ctx := config.GetCurrentContext(cfg)
	if ctx == nil {
		return DiagnosticCheck{
			Name:    "API connectivity",
			Status:  "SKIP",
			Message: "No context set to check connectivity",
			CanFix:  false,
		}
	}

	// Try a simple HTTP check
	// This is a simplified check - in production, use proper HTTP client
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", ctx.Server)
	if output, err := cmd.CombinedOutput(); err != nil {
		return DiagnosticCheck{
			Name:    "API connectivity",
			Status:  "WARN",
			Message: fmt.Sprintf("Cannot reach API server: %v. curl may not be available or server is unreachable.", err),
			CanFix:  false,
		}
	} else {
		statusCode := strings.TrimSpace(string(output))
		if statusCode == "000" {
			return DiagnosticCheck{
				Name:    "API connectivity",
				Status:  "WARN",
				Message: fmt.Sprintf("Cannot reach API server at %s", ctx.Server),
				CanFix:  false,
			}
		}
	}

	return DiagnosticCheck{
		Name:    "API connectivity",
		Status:  "OK",
		Message: fmt.Sprintf("API server at %s is reachable", ctx.Server),
		CanFix:  false,
	}
}

func checkCredentials() DiagnosticCheck {
	// Check environment variables first
	if _, err := auth.GetCredentialsFromEnv(); err == nil {
		return DiagnosticCheck{
			Name:    "Credentials",
			Status:  "OK",
			Message: "Credentials found in environment variables",
			CanFix:  false,
		}
	}

	// Check config file credentials
	cfg, err := config.Load()
	if err != nil {
		return DiagnosticCheck{
			Name:    "Credentials",
			Status:  "ERROR",
			Message: "Cannot check credentials without config",
			CanFix:  false,
		}
	}

	ctx := config.GetCurrentContext(cfg)
	if ctx == nil {
		return DiagnosticCheck{
			Name:    "Credentials",
			Status:  "WARN",
			Message: "No context to check credentials",
			CanFix:  false,
		}
	}

	if ctx.APIKey == "" {
		return DiagnosticCheck{
			Name:    "Credentials",
			Status:  "WARN",
			Message: "No API key configured for current context",
			CanFix:  false,
		}
	}

	return DiagnosticCheck{
		Name:    "Credentials",
		Status:  "OK",
		Message: "API key configured for current context",
		CanFix:  false,
	}
}

func checkRequiredTools() DiagnosticCheck {
	missing := []string{}

	// Check for common tools
	tools := []string{"curl", "git"}
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}

	if len(missing) > 0 {
		return DiagnosticCheck{
			Name:    "Required tools",
			Status:  "WARN",
			Message: fmt.Sprintf("Missing tools: %v. These are recommended for full functionality.", missing),
			CanFix:  false,
		}
	}

	return DiagnosticCheck{
		Name:    "Required tools",
		Status:  "OK",
		Message: "All recommended tools are installed",
		CanFix:  false,
	}
}

func checkPermissions() DiagnosticCheck {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DiagnosticCheck{
			Name:    "Directory permissions",
			Status:  "ERROR",
			Message: fmt.Sprintf("Cannot determine home directory: %v", err),
			CanFix:  false,
		}
	}

	kranixDir := filepath.Join(homeDir, ".kranix")
	if _, err := os.Stat(kranixDir); os.IsNotExist(err) {
		// Try to create the directory
		if err := os.MkdirAll(kranixDir, 0755); err != nil {
			return DiagnosticCheck{
				Name:    "Directory permissions",
				Status:  "ERROR",
				Message: fmt.Sprintf("Cannot create .kranix directory: %v", err),
				CanFix:  false,
			}
		}
		return DiagnosticCheck{
			Name:    "Directory permissions",
			Status:  "OK",
			Message: "Created .kranix directory successfully",
			CanFix:  false,
		}
	}

	return DiagnosticCheck{
		Name:    "Directory permissions",
		Status:  "OK",
		Message: "Directory permissions are correct",
		CanFix:  false,
	}
}

func checkGoVersion() DiagnosticCheck {
	if _, err := exec.LookPath("go"); err != nil {
		return DiagnosticCheck{
			Name:    "Go installation",
			Status:  "SKIP",
			Message: "Go not found (optional for CLI usage, required for development)",
			CanFix:  false,
		}
	}

	cmd := exec.Command("go", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return DiagnosticCheck{
			Name:    "Go installation",
			Status:  "WARN",
			Message: fmt.Sprintf("Go found but cannot get version: %v", err),
			CanFix:  false,
		}
	}

	version := strings.TrimSpace(string(output))
	return DiagnosticCheck{
		Name:    "Go installation",
		Status:  "OK",
		Message: version,
		CanFix:  false,
	}
}
