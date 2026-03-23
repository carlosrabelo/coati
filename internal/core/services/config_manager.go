package services

import (
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"

	"coati/internal/core/domain"
	"coati/internal/core/ports"
)

const (
	defaultConfigDir = "/etc/coati"
	defaultAppConfig = "/etc/coati/config.yaml"
)

type ConfigManager struct {
	gistFetcher ports.GistFetcher
	logger      *slog.Logger
}

func NewConfigManager(logger *slog.Logger, gistFetcher ports.GistFetcher) *ConfigManager {
	return &ConfigManager{
		gistFetcher: gistFetcher,
		logger:      logger,
	}
}

func (cm *ConfigManager) LoadConfig() (domain.AppConfig, error) {
	var config domain.AppConfig

	// Check if file exists
	if _, err := os.Stat(defaultAppConfig); os.IsNotExist(err) {
		cm.logger.Debug("App config file not found", "path", defaultAppConfig)
		return config, nil // Return empty config, not an error
	}

	cm.logger.Info("Loading app config", "path", defaultAppConfig)
	data, err := os.ReadFile(defaultAppConfig)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	cm.logger.Debug("Config loaded", "gist_id", config.GistID, "has_token", config.GitHubToken != "")
	return config, nil
}

func (cm *ConfigManager) SaveConfig(gistID, token string) error {
	if gistID == "" && token == "" {
		return fmt.Errorf("no gist-id or token provided to save")
	}

	newConfig := domain.AppConfig{
		GistID:      gistID,
		GitHubToken: token,
	}

	data, err := yaml.Marshal(newConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.MkdirAll(defaultConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(defaultAppConfig, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	cm.logger.Info("Configuration saved securely", "path", defaultAppConfig)
	return nil
}

func (cm *ConfigManager) FetchGist(gistID, token string) ([]byte, error) {
	if gistID == "" {
		return nil, fmt.Errorf("gist ID is empty")
	}
	if token == "" {
		return nil, fmt.Errorf("github token is empty")
	}

	cm.logger.Info("Fetching configuration from Gist", "gist_id", gistID)
	return cm.gistFetcher.Fetch(gistID, token)
}
