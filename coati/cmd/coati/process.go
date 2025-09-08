package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"coati/coati/internal/app"
	"coati/coati/internal/core/services"
)

func newProcessCmd() *cobra.Command {
	var cfg app.Config

	cmd := &cobra.Command{
		Use:   "process",
		Short: "Process YAML configuration and write hosts and SSH config files",
		Run: func(cmd *cobra.Command, args []string) {
			opts := &slog.HandlerOptions{Level: slog.LevelInfo}
			if cfg.Verbose {
				opts.Level = slog.LevelDebug
			}
			logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

			application, err := app.New(logger, cfg)
			if err != nil {
				logger.Error("Failed to initialize application", "error", err)
				os.Exit(1)
			}
			if err := application.Run(); err != nil {
				logger.Error("Application error", "error", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&cfg.HostsListFile, "hosts-list", services.HostsYAMLPath(), "Input YAML configuration file")
	cmd.Flags().StringVar(&cfg.OutputHostsFile, "output-hosts", services.DefaultHostsOutput(), "Output hosts file")
	cmd.Flags().StringVar(&cfg.OutputConfigFile, "output-config", "", "Output SSH config file (default: ~/.ssh/config for each target user)")
	cmd.Flags().StringVar(&cfg.HostsTemplateFile, "hosts-template", "", "Path to custom hosts template file (uses embedded default if not specified)")
	cmd.Flags().StringVar(&cfg.GistID, "gist-id", "", "GitHub Gist ID to fetch config from")
	cmd.Flags().StringVar(&cfg.GistFile, "gist-file", "", "Filename inside the Gist to use (uses first file if not specified)")
	cmd.Flags().StringVar(&cfg.GitHubToken, "github-token", "", "GitHub Personal Access Token (or via GITHUB_TOKEN env)")
	cmd.Flags().BoolVar(&cfg.DryRun, "dry-run", false, "Print generated content to stdout without writing to files")
	cmd.Flags().BoolVar(&cfg.Check, "check", false, "Show diff between current files and what would be written, without writing")
	cmd.Flags().BoolVar(&cfg.Merge, "merge", false, "Preserve existing file content in ORIGINAL section, manage COATI section only")
	cmd.Flags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose logging")
	cmd.Flags().BoolVarP(&cfg.ForceRefresh, "force-refresh", "f", false, "Force refresh of Gist configuration (bypass cache)")
	cmd.Flags().BoolVar(&cfg.Backup, "backup", true, "Create a backup (.bak) of files before writing")
	cmd.Flags().BoolVar(&cfg.AllowUnsafeHooks, "allow-unsafe-hooks", false, "Bypass command allowlist for post-hooks (potentially unsafe)")
	cmd.Flags().StringSliceVar(&cfg.AllowedHooks, "allowed-hooks", []string{}, "Comma-separated list of additional allowed commands for post-hooks")
	cmd.Flags().StringVar(&cfg.Scope, "scope", "", "What to update: system, user, or all (default: all as root, user otherwise)")
	cmd.Flags().StringSliceVar(&cfg.Users, "users", []string{}, "Users whose ~/.ssh/config to update (default: SUDO_USER or current user)")

	return cmd
}
