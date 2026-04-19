package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"coati/coati/internal/adapters/secondary"
	"coati/coati/internal/core/domain"
	"coati/coati/internal/core/services"
	"coati/coati/internal/templates"
)

const DefaultHostsInput = "/etc/coati/hosts.yaml"

type Config struct {
	HostsListFile     string
	OutputHostsFile   string
	OutputConfigFile  string
	HostsTemplateFile string
	GistID            string
	GistFile          string
	GitHubToken       string
	SaveConfig        bool
	DryRun            bool
	Check             bool
	Merge             bool
	Verbose           bool
	ForceRefresh      bool
}

type Application struct {
	logger        *slog.Logger
	cfg           Config
	configManager *services.ConfigManager
	fsAdapter     *secondary.FSAdapter
}

func New(logger *slog.Logger, cfg Config) (*Application, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseFetcher := secondary.NewGistFetcher()
	cacheDir := filepath.Join(home, ".coati", "cache", "gists")

	ttl := 5 * time.Minute
	if cfg.ForceRefresh {
		ttl = 0
	}

	cachedFetcher := services.NewCachedGistFetcher(baseFetcher, cacheDir, ttl)

	return &Application{
		logger:        logger,
		cfg:           cfg,
		configManager: services.NewConfigManager(logger, cachedFetcher),
		fsAdapter:     secondary.NewFSAdapter(),
	}, nil
}

func (app *Application) Run() error {
	// 1. Load App Config
	appConfig, err := app.configManager.LoadConfig()
	if err != nil {
		app.logger.Warn("Failed to load app config", "error", err)
	}

	// 2. Override with CLI / Env
	// Priority: CLI flag > local hosts file > AppConfig Gist ID
	if app.cfg.GistID == "" && app.cfg.HostsListFile == DefaultHostsInput {
		app.cfg.GistID = appConfig.GistID
	}
	if app.cfg.GitHubToken == "" {
		app.cfg.GitHubToken = appConfig.GitHubToken
	}
	if app.cfg.GitHubToken == "" {
		app.cfg.GitHubToken = os.Getenv("GITHUB_TOKEN")
	}

	// 3. Handle Save Config
	if app.cfg.SaveConfig {
		if err := app.configManager.SaveConfig(app.cfg.GistID, app.cfg.GitHubToken); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		return nil
	}

	// 4. Create Metrics
	metrics := domain.NewMetrics()
	defer func() {
		if app.cfg.Verbose {
			app.logger.Info("Performance Report", "metrics", metrics.Report())
		}
	}()

	// 5. Resolve input path (Gist or local)
	startFetch := time.Now()
	inputPath, cleanup, err := app.resolveInputPath()
	metrics.FetchDuration = time.Since(startFetch)
	defer cleanup()
	if err != nil {
		return err
	}

	// 6. Read & Parse Config
	startParse := time.Now()
	data, err := app.fsAdapter.ReadFile(inputPath)
	if err != nil {
		app.logger.Error("Failed to read input file", "path", inputPath, "error", err)
		if app.cfg.GistID == "" {
			fmt.Println("\nTip: Coati is not configured. Run with --save-config to setup Gist or ensure local file exists.")
		}
		return err
	}
	globalConfig, err := loadGlobalConfig(data)
	if err != nil {
		return err
	}
	metrics.ParseDuration = time.Since(startParse)

	// 7. Generate Content
	hostsGenerator := services.NewHostsGenerator(globalConfig, app.fsAdapter, app.resolveTemplate())
	sshGenerator := services.NewSSHGenerator(globalConfig)

	app.logger.Info("Generating hosts file...")
	startHosts := time.Now()
	hostsData, err := hostsGenerator.GenerateHosts()
	metrics.HostsGenDuration = time.Since(startHosts)
	if err != nil {
		return fmt.Errorf("failed to generate hosts: %w", err)
	}

	app.logger.Info("Generating SSH config...")
	startSSH := time.Now()
	sshData, err := sshGenerator.GenerateSSHConfig()
	metrics.SSHGenDuration = time.Since(startSSH)
	if err != nil {
		return fmt.Errorf("failed to generate SSH config: %w", err)
	}

	// 8. Dry Run: show full generated content without merging
	if app.cfg.DryRun {
		fmt.Println("\n=== DRY RUN MODE: No files will be modified ===")
		printColoredDiff(fmt.Sprintf("Content for %s", app.cfg.OutputHostsFile), hostsData)
		printColoredDiff(fmt.Sprintf("Content for %s", app.cfg.OutputConfigFile), sshData)
		fmt.Println("\n=== End of Dry Run ===")
		return nil
	}

	// 9. Merge: wrap existing content and embed new content inside markers
	if app.cfg.Merge {
		hostsData, err = mergeWithMarkers(app.cfg.OutputHostsFile, hostsData)
		if err != nil {
			return fmt.Errorf("failed to merge hosts file: %w", err)
		}
		sshData, err = mergeWithMarkers(app.cfg.OutputConfigFile, sshData)
		if err != nil {
			return fmt.Errorf("failed to merge SSH config: %w", err)
		}
	}

	// 10. Check: show diff between current files and what would be written
	if app.cfg.Check {
		fmt.Println("\n=== CHECK MODE: No files will be modified ===")
		printFileDiff(app.cfg.OutputHostsFile, app.cfg.OutputHostsFile, hostsData)
		printFileDiff(app.cfg.OutputConfigFile, app.cfg.OutputConfigFile, sshData)
		fmt.Println("\n=== End of Check ===")
		return nil
	}

	app.logger.Info("Writing hosts file", "path", app.cfg.OutputHostsFile)
	if err := app.fsAdapter.WriteFile(app.cfg.OutputHostsFile, hostsData, 0644); err != nil {
		return fmt.Errorf("failed to write hosts file: %w", err)
	}

	app.logger.Info("Writing SSH config", "path", app.cfg.OutputConfigFile)
	if err := app.fsAdapter.WriteFile(app.cfg.OutputConfigFile, sshData, 0600); err != nil {
		return fmt.Errorf("failed to write SSH config: %w", err)
	}

	// 11. Post-Execution Hooks
	if len(globalConfig.PostHooks) > 0 {
		app.logger.Info("Running post-execution hooks...")
		return app.runHooks(globalConfig.PostHooks)
	}

	app.logger.Info("Coati execution completed successfully")
	return nil
}

// resolveInputPath returns the path to the configuration file and a cleanup
// function to be deferred. For local files, cleanup is a no-op.
func (app *Application) resolveInputPath() (string, func(), error) {
	noop := func() {}

	if app.cfg.GistID == "" {
		app.logger.Info("Using local configuration file", "path", app.cfg.HostsListFile)
		return app.cfg.HostsListFile, noop, nil
	}

	if app.cfg.GitHubToken == "" {
		app.logger.Error("GitHub token required for Gist fetch")
		fmt.Println("Tip: Provide --github-token or set GITHUB_TOKEN env var.")
		return "", noop, fmt.Errorf("github token missing")
	}

	content, err := app.configManager.FetchGist(app.cfg.GistID, app.cfg.GitHubToken, app.cfg.GistFile)
	if err != nil {
		return "", noop, fmt.Errorf("failed to fetch Gist: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "hosts-*.yaml")
	if err != nil {
		return "", noop, fmt.Errorf("failed to create temp file: %w", err)
	}
	cleanup := func() { os.Remove(tmpFile.Name()) }

	if _, err := tmpFile.Write(content); err != nil {
		cleanup()
		return "", noop, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	app.logger.Info("Using Gist configuration", "gist_id", app.cfg.GistID, "temp_file", tmpFile.Name())
	return tmpFile.Name(), cleanup, nil
}

// loadGlobalConfig parses and validates a YAML configuration.
func loadGlobalConfig(data []byte) (domain.GlobalConfig, error) {
	var cfg domain.GlobalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse YAML configuration: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("invalid configuration: %w", err)
	}
	return cfg, nil
}

// resolveTemplate returns the hosts template content. Falls back to the
// embedded default if no custom template file is specified or readable.
func (app *Application) resolveTemplate() string {
	if app.cfg.HostsTemplateFile == "" {
		return templates.HostsTemplate
	}
	data, err := app.fsAdapter.ReadFile(app.cfg.HostsTemplateFile)
	if err != nil {
		app.logger.Warn("Failed to read custom template file. Using default.", "path", app.cfg.HostsTemplateFile, "error", err)
		return templates.HostsTemplate
	}
	return string(data)
}
