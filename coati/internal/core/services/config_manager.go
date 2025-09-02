package services

import (
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"

	"coati/coati/internal/core/domain"
	"coati/coati/internal/core/ports"
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

	appConfig := AppConfigPath()

	if _, err := os.Stat(appConfig); os.IsNotExist(err) {
		cm.logger.Debug("App config file not found", "path", appConfig)
		return config, nil
	}

	cm.logger.Info("Loading app config", "path", appConfig)
	data, err := os.ReadFile(appConfig)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	cm.logger.Debug("Config loaded", "gist_id", config.GistID, "has_token", config.GitHubToken != "")
	return config, nil
}

func (cm *ConfigManager) FetchGist(gistID, token, gistFile string) ([]byte, error) {
	if gistID == "" {
		return nil, fmt.Errorf("gist ID is empty")
	}
	if token == "" {
		return nil, fmt.Errorf("github token is empty")
	}

	cm.logger.Info("Fetching configuration from Gist", "gist_id", gistID)
	return cm.gistFetcher.Fetch(gistID, token, gistFile)
}
