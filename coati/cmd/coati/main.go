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

	var applyCmd = &cobra.Command{
		Use:   "apply",
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

	applyCmd.Flags().StringVar(&cfg.HostsListFile, "hosts-list", app.DefaultHostsInput, "Input YAML configuration file")
	applyCmd.Flags().StringVar(&cfg.OutputHostsFile, "output-hosts", "data/out/etc/hosts", "Output hosts file")
	applyCmd.Flags().StringVar(&cfg.OutputConfigFile, "output-config", "data/out/ssh/config", "Output SSH config file")
	applyCmd.Flags().StringVar(&cfg.HostsTemplateFile, "hosts-template", "", "Path to custom hosts template file (uses embedded default if not specified)")
	applyCmd.Flags().StringVar(&cfg.GistID, "gist-id", "", "GitHub Gist ID to fetch config from")
	applyCmd.Flags().StringVar(&cfg.GistFile, "gist-file", "", "Filename inside the Gist to use (uses first file if not specified)")
	applyCmd.Flags().StringVar(&cfg.GitHubToken, "github-token", "", "GitHub Personal Access Token (or via GITHUB_TOKEN env)")
	applyCmd.Flags().BoolVar(&cfg.SaveConfig, "save-config", false, "Save gist-id and token to secure config file")
	applyCmd.Flags().BoolVar(&cfg.DryRun, "dry-run", false, "Print generated content to stdout without writing to files")
	applyCmd.Flags().BoolVar(&cfg.Check, "check", false, "Show diff between current files and what would be written, without writing")
	applyCmd.Flags().BoolVar(&cfg.Merge, "merge", false, "Preserve existing file content in ORIGINAL section, manage COATI section only")
	applyCmd.Flags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose logging")
	applyCmd.Flags().BoolVarP(&cfg.ForceRefresh, "force-refresh", "f", false, "Force refresh of Gist configuration (bypass cache)")

	var rootCmd = &cobra.Command{
		Use:   "coati",
		Short: "Manage /etc/hosts and ~/.ssh/config from a YAML definition",
	}

	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(completionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
