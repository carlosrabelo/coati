package app

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"coati/coati/internal/adapters/secondary"
	"coati/coati/internal/core/domain"
	"coati/coati/internal/core/services"
	"coati/coati/internal/templates"
)

const (
	ScopeSystem = "system"
	ScopeUser   = "user"
	ScopeAll    = "all"
)

type Config struct {
	HostsListFile     string
	OutputHostsFile   string
	OutputConfigFile  string
	HostsTemplateFile string
	GistID            string
	GistFile          string
	GitHubToken       string
	DryRun            bool
	Check             bool
	Merge             bool
	Verbose           bool
	ForceRefresh      bool
	Backup            bool
	AllowUnsafeHooks  bool
	AllowedHooks      []string
	Scope             string
	Users             []string
}

type Application struct {
	logger        *slog.Logger
	cfg           Config
	configManager *services.ConfigManager
	fsAdapter     *secondary.FSAdapter
}

func New(logger *slog.Logger, cfg Config) (*Application, error) {
	baseFetcher := secondary.NewGistFetcher()
	cacheDir := services.GistCacheDir()

	ttl := 5 * time.Minute
	if cfg.ForceRefresh {
		ttl = 0
	}

	cachedFetcher := services.NewCachedGistFetcher(baseFetcher, cacheDir, ttl)

	if cfg.HostsListFile == "" {
		cfg.HostsListFile = services.HostsYAMLPath()
	}
	if cfg.OutputHostsFile == "" {
		cfg.OutputHostsFile = services.DefaultHostsOutput()
	}

	return &Application{
		logger:        logger,
		cfg:           cfg,
		configManager: services.NewConfigManager(logger, cachedFetcher),
		fsAdapter:     secondary.NewFSAdapter(),
	}, nil
}

func (app *Application) Run() error {
	scope, err := normalizeScope(app.cfg.Scope)
	if err != nil {
		return err
	}
	scope, err = app.resolveEffectiveScope(scope)
	if err != nil {
		return err
	}
	app.cfg.Scope = scope

	appConfig, err := app.configManager.LoadConfig()
	if err != nil {
		app.logger.Warn("Failed to load app config", "error", err)
	}

	if app.cfg.GitHubToken == "" {
		app.cfg.GitHubToken = appConfig.GitHubToken
	}
	if app.cfg.GitHubToken == "" {
		app.cfg.GitHubToken = os.Getenv("GITHUB_TOKEN")
	}

	allowedHooks := app.cfg.AllowedHooks
	for _, h := range appConfig.AllowedHooks {
		found := false
		for _, existing := range allowedHooks {
			if existing == h {
				found = true
				break
			}
		}
		if !found {
			allowedHooks = append(allowedHooks, h)
		}
	}
	app.cfg.AllowedHooks = allowedHooks

	metrics := domain.NewMetrics()
	defer func() {
		if app.cfg.Verbose {
			app.logger.Info("Performance Report", "metrics", metrics.Report())
		}
	}()

	startFetch := time.Now()
	inputPath, cleanup, err := app.resolveInputPath()
	metrics.FetchDuration = time.Since(startFetch)
	defer cleanup()
	if err != nil {
		return err
	}

	startParse := time.Now()
	data, err := app.fsAdapter.ReadFile(inputPath)
	if err != nil {
		app.logger.Error("Failed to read input file", "path", inputPath, "error", err)
		if app.cfg.GistID == "" {
			fmt.Printf("\nTip: Coati is not configured. Create %s or ensure %s exists.\n",
				services.AppConfigPath(), services.HostsYAMLPath())
		}
		return err
	}
	globalConfig, err := loadGlobalConfig(data)
	if err != nil {
		return err
	}
	metrics.ParseDuration = time.Since(startParse)

	hostsGenerator := services.NewHostsGenerator(globalConfig, app.fsAdapter, app.resolveTemplate())
	sshGenerator := services.NewSSHGenerator(globalConfig)

	var hostsData []byte
	if scope == ScopeSystem || scope == ScopeAll {
		app.logger.Info("Generating hosts file...")
		startHosts := time.Now()
		hostsData, err = hostsGenerator.GenerateHosts()
		metrics.HostsGenDuration = time.Since(startHosts)
		if err != nil {
			return fmt.Errorf("failed to generate hosts: %w", err)
		}
	}

	var sshData []byte
	if scope == ScopeUser || scope == ScopeAll {
		app.logger.Info("Generating SSH config...")
		startSSH := time.Now()
		sshData, err = sshGenerator.GenerateSSHConfig()
		metrics.SSHGenDuration = time.Since(startSSH)
		if err != nil {
			return fmt.Errorf("failed to generate SSH config: %w", err)
		}
	}

	targetUsers, err := services.ResolveTargetUsers(app.cfg.Users)
	if err != nil {
		return err
	}
	sshTargets, err := app.resolveSSHTargets(targetUsers)
	if err != nil {
		return err
	}

	if app.cfg.DryRun {
		fmt.Println("\n=== DRY RUN MODE: No files will be modified ===")
		if hostsData != nil {
			printColoredDiff(fmt.Sprintf("Content for %s", app.cfg.OutputHostsFile), hostsData)
		}
		if sshData != nil {
			for _, target := range sshTargets {
				printColoredDiff(fmt.Sprintf("Content for %s", target.path), sshData)
			}
		}
		fmt.Println("\n=== End of Dry Run ===")
		return nil
	}

	if app.cfg.Merge {
		if hostsData != nil {
			hostsData, err = mergeWithMarkers(app.cfg.OutputHostsFile, hostsData)
			if err != nil {
				return fmt.Errorf("failed to merge hosts file: %w", err)
			}
		}
		if sshData != nil {
			if len(sshTargets) == 1 {
				sshData, err = mergeWithMarkers(sshTargets[0].path, sshData)
				if err != nil {
					return fmt.Errorf("failed to merge SSH config: %w", err)
				}
			}
		}
	}

	if app.cfg.Check {
		fmt.Println("\n=== CHECK MODE: No files will be modified ===")
		if hostsData != nil {
			printFileDiff(app.cfg.OutputHostsFile, app.cfg.OutputHostsFile, hostsData)
		}
		if sshData != nil {
			for _, target := range sshTargets {
				content := sshData
				if app.cfg.Merge && len(sshTargets) > 1 {
					content, err = mergeWithMarkers(target.path, sshData)
					if err != nil {
						return fmt.Errorf("failed to merge SSH config for %s: %w", target.username, err)
					}
				}
				printFileDiff(target.path, target.path, content)
			}
		}
		fmt.Println("\n=== End of Check ===")
		return nil
	}

	var writeErrs []error

	if hostsData != nil {
		if app.cfg.Backup {
			if err := app.createBackup(app.cfg.OutputHostsFile); err != nil {
				app.logger.Warn("Failed to create backup for hosts file", "path", app.cfg.OutputHostsFile, "error", err)
			}
		}
		app.logger.Info("Writing hosts file", "path", app.cfg.OutputHostsFile)
		if err := writeAtomic(app.cfg.OutputHostsFile, hostsData, 0644); err != nil {
			if scope == ScopeAll && isPermissionError(err) {
				app.logger.Warn("Skipping hosts write (permission denied); run with sudo for --scope system",
					"path", app.cfg.OutputHostsFile, "error", err)
				fmt.Println("Tip: run with sudo for --scope system")
			} else if scope == ScopeSystem && isPermissionError(err) {
				return fmt.Errorf("failed to write hosts file %s (root required): %w", app.cfg.OutputHostsFile, err)
			} else {
				return fmt.Errorf("failed to write hosts file: %w", err)
			}
		}
	}

	if sshData != nil {
		for _, target := range sshTargets {
			content := sshData
			if app.cfg.Merge && len(sshTargets) > 1 {
				content, err = mergeWithMarkers(target.path, sshData)
				if err != nil {
					writeErrs = append(writeErrs, fmt.Errorf("merge SSH for %s: %w", target.username, err))
					continue
				}
			}

			if app.cfg.Backup {
				if err := app.createBackup(target.path); err != nil {
					app.logger.Warn("Failed to create backup for SSH config file", "path", target.path, "error", err)
				}
			}
			app.logger.Info("Writing SSH config", "path", target.path, "user", target.username)
			if err := writeAtomic(target.path, content, 0600); err != nil {
				writeErrs = append(writeErrs, fmt.Errorf("write SSH for %s: %w", target.username, err))
				continue
			}
			if err := chownPath(target.path, target.uid, target.gid); err != nil {
				app.logger.Warn("Failed to chown SSH config", "path", target.path, "user", target.username, "error", err)
			}
			if err := chownPath(filepath.Dir(target.path), target.uid, target.gid); err != nil {
				app.logger.Debug("Failed to chown SSH directory", "path", filepath.Dir(target.path), "error", err)
			}
		}
	}

	if len(writeErrs) > 0 {
		return errors.Join(writeErrs...)
	}

	if len(globalConfig.PostHooks) > 0 {
		app.logger.Info("Running post-execution hooks...")
		return app.runHooks(globalConfig.PostHooks)
	}

	app.logger.Info("Coati execution completed successfully")
	return nil
}

type sshTarget struct {
	username string
	path     string
	uid      int
	gid      int
}

func (app *Application) resolveSSHTargets(usernames []string) ([]sshTarget, error) {
	if app.cfg.OutputConfigFile != "" {
		uid, gid := -1, -1
		if u, err := user.Current(); err == nil {
			uid, _ = strconv.Atoi(u.Uid)
			gid, _ = strconv.Atoi(u.Gid)
		}
		return []sshTarget{{
			username: usernames[0],
			path:     app.cfg.OutputConfigFile,
			uid:      uid,
			gid:      gid,
		}}, nil
	}

	targets := make([]sshTarget, 0, len(usernames))
	for _, name := range usernames {
		u, err := user.Lookup(name)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup user %q: %w", name, err)
		}
		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			return nil, fmt.Errorf("invalid uid for %q: %w", name, err)
		}
		gid, err := strconv.Atoi(u.Gid)
		if err != nil {
			return nil, fmt.Errorf("invalid gid for %q: %w", name, err)
		}
		targets = append(targets, sshTarget{
			username: name,
			path:     services.SSHConfigPath(u.HomeDir),
			uid:      uid,
			gid:      gid,
		})
	}
	return targets, nil
}

func normalizeScope(scope string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "":
		if services.RunningAsRoot() {
			return ScopeAll, nil
		}
		return ScopeUser, nil
	case ScopeAll:
		return ScopeAll, nil
	case ScopeSystem:
		return ScopeSystem, nil
	case ScopeUser:
		return ScopeUser, nil
	default:
		return "", fmt.Errorf("invalid scope %q (want system, user, or all)", scope)
	}
}

func (app *Application) resolveEffectiveScope(scope string) (string, error) {
	if services.RunningAsRoot() || !isSystemHostsPath(app.cfg.OutputHostsFile) {
		return scope, nil
	}

	switch scope {
	case ScopeSystem, ScopeAll:
		return "", fmt.Errorf("--scope %s requires root to write %s; run with sudo (or use --scope user)",
			scope, app.cfg.OutputHostsFile)
	default:
		return scope, nil
	}
}

func isSystemHostsPath(path string) bool {
	return path == "" || path == services.DefaultHostsOutput()
}

func isPermissionError(err error) bool {
	return errors.Is(err, os.ErrPermission) || os.IsPermission(err)
}

func chownPath(path string, uid, gid int) error {
	if uid < 0 || gid < 0 {
		return nil
	}
	if os.Geteuid() != 0 {
		return nil
	}
	return os.Chown(path, uid, gid)
}

func (app *Application) resolveInputPath() (string, func(), error) {
	noop := func() {}
	app.logger.Info("Using local configuration file", "path", app.cfg.HostsListFile)
	return app.cfg.HostsListFile, noop, nil
}

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

func WriteAtomicPublic(dest string, data []byte, perm os.FileMode) error {
	return writeAtomic(dest, data, perm)
}

func writeAtomic(dest string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".coati-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { os.Remove(tmpName) }()

	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, dest)
}

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

func (app *Application) createBackup(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to stat file for backup: %w", err)
	}

	backupPath := path + ".bak"
	app.logger.Debug("Creating backup copy", "path", path, "backupPath", backupPath)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file for backup: %w", err)
	}

	if err := writeAtomic(backupPath, data, info.Mode().Perm()); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}
