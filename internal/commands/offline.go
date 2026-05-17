package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kranix-io/kranix-cli/internal/client"
	"github.com/spf13/cobra"
)

var offlineEnable bool
var offlineDisable bool
var offlineStatus bool
var offlineSync bool

var offlineCmd = &cobra.Command{
	Use:   "offline",
	Short: "Manage offline/air-gap mode",
	Long:  "Enable or disable offline mode for basic operations when kranix-api is not reachable. Caches workloads locally for read-only access.",
	Example: `  kranix offline --enable
  kranix offline --status
  kranix offline --sync`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if offlineEnable {
			return enableOfflineMode()
		}
		if offlineDisable {
			return disableOfflineMode()
		}
		if offlineStatus {
			return showOfflineStatus()
		}
		if offlineSync {
			return syncOfflineCache()
		}
		return cmd.Help()
	},
}

func init() {
	offlineCmd.Flags().BoolVar(&offlineEnable, "enable", false, "Enable offline mode")
	offlineCmd.Flags().BoolVar(&offlineDisable, "disable", false, "Disable offline mode")
	offlineCmd.Flags().BoolVar(&offlineStatus, "status", false, "Show offline mode status")
	offlineCmd.Flags().BoolVar(&offlineSync, "sync", false, "Sync offline cache with API")
}

const offlineCacheDir = ".kranix/cache"
const offlineConfigFile = ".kranix/offline_config.json"

type OfflineConfig struct {
	Enabled    bool      `json:"enabled"`
	LastSync   time.Time `json:"last_sync"`
	CacheDir   string    `json:"cache_dir"`
	Namespaces []string  `json:"namespaces"`
}

func enableOfflineMode() error {
	config := &OfflineConfig{
		Enabled:    true,
		LastSync:   time.Now(),
		CacheDir:   offlineCacheDir,
		Namespaces: []string{"default"},
	}

	// Create cache directory
	if err := os.MkdirAll(offlineCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Save config
	configDir := filepath.Dir(offlineConfigFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(offlineConfigFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Println("Offline mode enabled")
	fmt.Printf("Cache directory: %s\n", offlineCacheDir)
	fmt.Println("Use 'kranix offline --sync' to populate cache")

	return nil
}

func disableOfflineMode() error {
	if _, err := os.Stat(offlineConfigFile); os.IsNotExist(err) {
		return fmt.Errorf("offline mode is not enabled")
	}

	if err := os.Remove(offlineConfigFile); err != nil {
		return fmt.Errorf("failed to disable offline mode: %w", err)
	}

	fmt.Println("Offline mode disabled")
	return nil
}

func showOfflineStatus() error {
	config, err := loadOfflineConfig()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Offline mode: disabled")
			return nil
		}
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Offline mode: enabled")
	fmt.Printf("Last sync: %s\n", config.LastSync.Format(time.RFC3339))
	fmt.Printf("Cache directory: %s\n", config.CacheDir)
	fmt.Printf("Tracked namespaces: %v\n", config.Namespaces)

	// Show cache statistics
	files, err := os.ReadDir(offlineCacheDir)
	if err == nil {
		fmt.Printf("Cached workloads: %d\n", len(files))
	}

	return nil
}

func syncOfflineCache() error {
	config, err := loadOfflineConfig()
	if err != nil {
		return fmt.Errorf("offline mode not enabled: %w", err)
	}

	_, creds, err := getCredentials()
	if err != nil {
		return fmt.Errorf("failed to get credentials: %w", err)
	}

	cli := client.New(creds.Server, creds.APIKey)

	fmt.Println("Syncing offline cache...")

	// Sync workloads for each namespace
	for _, ns := range config.Namespaces {
		workloads, err := cli.ListWorkloads(context.TODO(), ns)
		if err != nil {
			fmt.Printf("Warning: failed to sync namespace %s: %v\n", ns, err)
			continue
		}

		for _, w := range workloads {
			cacheFile := filepath.Join(offlineCacheDir, fmt.Sprintf("%s_%s.json", ns, w.Name))
			data, err := json.MarshalIndent(w, "", "  ")
			if err != nil {
				fmt.Printf("Warning: failed to cache workload %s: %v\n", w.Name, err)
				continue
			}

			if err := os.WriteFile(cacheFile, data, 0644); err != nil {
				fmt.Printf("Warning: failed to write cache for %s: %v\n", w.Name, err)
				continue
			}
		}
		fmt.Printf("Synced namespace: %s\n", ns)
	}

	// Update last sync time
	config.LastSync = time.Now()
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	if err := os.WriteFile(offlineConfigFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Println("Sync completed successfully")
	return nil
}

func loadOfflineConfig() (*OfflineConfig, error) {
	data, err := os.ReadFile(offlineConfigFile)
	if err != nil {
		return nil, err
	}

	var config OfflineConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// IsOfflineModeEnabled checks if offline mode is currently enabled
func IsOfflineModeEnabled() bool {
	config, err := loadOfflineConfig()
	if err != nil {
		return false
	}
	return config.Enabled
}

// GetCachedWorkloads retrieves workloads from offline cache
func GetCachedWorkloads(namespace string) ([]*client.WorkloadStatus, error) {
	var workloads []*client.WorkloadStatus

	files, err := os.ReadDir(offlineCacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(offlineCacheDir, f.Name()))
		if err != nil {
			continue
		}

		var w client.WorkloadStatus
		if err := json.Unmarshal(data, &w); err != nil {
			continue
		}

		if namespace == "" || w.Namespace == namespace {
			workloads = append(workloads, &w)
		}
	}

	return workloads, nil
}
