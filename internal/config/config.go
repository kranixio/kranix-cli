package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	configDir      = ".kranix"
	configFile     = "config"
	defaultServer  = "http://localhost:8080"
	defaultTimeout = "5m"
)

type Config struct {
	CurrentContext string    `yaml:"current-context"`
	Contexts       []Context `yaml:"contexts"`
	Defaults       Defaults  `yaml:"defaults"`
}

type Context struct {
	Name   string `yaml:"name"`
	Server string `yaml:"server"`
	APIKey string `yaml:"api-key"`
}

type Defaults struct {
	Namespace string `yaml:"namespace"`
	Output    string `yaml:"output"`
	Timeout   string `yaml:"timeout"`
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, configDir, configFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDirPath := filepath.Join(home, configDir)
	if err := os.MkdirAll(configDirPath, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDirPath, configFile)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func DefaultConfig() *Config {
	return &Config{
		CurrentContext: "default",
		Contexts: []Context{
			{
				Name:   "default",
				Server: defaultServer,
			},
		},
		Defaults: Defaults{
			Namespace: "default",
			Output:    "table",
			Timeout:   defaultTimeout,
		},
	}
}

func GetContext(cfg *Config, name string) *Context {
	for _, ctx := range cfg.Contexts {
		if ctx.Name == name {
			return &ctx
		}
	}
	return nil
}

func GetCurrentContext(cfg *Config) *Context {
	if cfg.CurrentContext == "" {
		return nil
	}
	return GetContext(cfg, cfg.CurrentContext)
}

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, configFile), nil
}
