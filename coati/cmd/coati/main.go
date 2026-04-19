package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"coati/coati/internal/app"
)

func main() {
	var cfg app.Config

	var rootCmd = &cobra.Command{
		Use:   "coati",
		Short: "Generate hosts and ssh config files",
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

	rootCmd.AddCommand(completionCmd)

	rootCmd.Flags().StringVar(&cfg.HostsListFile, "hosts-list", app.DefaultHostsInput, "Input YAML configuration file")
	rootCmd.Flags().StringVar(&cfg.OutputHostsFile, "output-hosts", "data/out/etc/hosts", "Output hosts file")
	rootCmd.Flags().StringVar(&cfg.OutputConfigFile, "output-config", "data/out/ssh/config", "Output SSH config file")
	rootCmd.Flags().StringVar(&cfg.HostsTemplateFile, "hosts-template", "", "Path to custom hosts template file (uses embedded default if not specified)")
	rootCmd.Flags().StringVar(&cfg.GistID, "gist-id", "", "GitHub Gist ID to fetch config from")
	rootCmd.Flags().StringVar(&cfg.GitHubToken, "github-token", "", "GitHub Personal Access Token (or via GITHUB_TOKEN env)")
	rootCmd.Flags().BoolVar(&cfg.SaveConfig, "save-config", false, "Save gist-id and token to secure config file")
	rootCmd.Flags().BoolVar(&cfg.DryRun, "dry-run", false, "Print generated content to stdout without writing to files")
	rootCmd.Flags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().BoolVarP(&cfg.ForceRefresh, "force-refresh", "f", false, "Force refresh of Gist configuration (bypass cache)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
