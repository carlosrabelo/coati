package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"coati/coati/internal/adapters/secondary"
	"coati/coati/internal/app"
	"coati/coati/internal/core/services"
)

const defaultLocalConfig = "data/src/gist.txt"

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Download Gist content and save to local config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		gistID, _ := cmd.Flags().GetString("gist-id")
		gistFile, _ := cmd.Flags().GetString("gist-file")
		token, _ := cmd.Flags().GetString("github-token")
		output, _ := cmd.Flags().GetString("output")
		verbose, _ := cmd.Flags().GetBool("verbose")

		logger := newLogger(verbose)

		token, gistID = resolveCredentials(token, gistID, logger)
		if gistID == "" {
			return fmt.Errorf("gist-id required (use --gist-id or save-config)")
		}
		if token == "" {
			return fmt.Errorf("github token required (use --github-token or GITHUB_TOKEN env)")
		}

		fetcher := services.NewCachedGistFetcher(secondary.NewGistFetcher(), cacheDir(), 0)
		content, err := fetcher.Fetch(gistID, token, gistFile)
		if err != nil {
			return fmt.Errorf("failed to fetch gist: %w", err)
		}

		if err := app.WriteAtomicPublic(output, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", output, err)
		}

		logger.Info("Gist saved", "path", output)
		return nil
	},
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Upload local config file to Gist",
	RunE: func(cmd *cobra.Command, args []string) error {
		gistID, _ := cmd.Flags().GetString("gist-id")
		gistFile, _ := cmd.Flags().GetString("gist-file")
		token, _ := cmd.Flags().GetString("github-token")
		input, _ := cmd.Flags().GetString("input")
		verbose, _ := cmd.Flags().GetBool("verbose")

		logger := newLogger(verbose)

		token, gistID = resolveCredentials(token, gistID, logger)
		if gistID == "" {
			return fmt.Errorf("gist-id required (use --gist-id or save-config)")
		}
		if token == "" {
			return fmt.Errorf("github token required (use --github-token or GITHUB_TOKEN env)")
		}

		content, err := os.ReadFile(input)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", input, err)
		}

		pusher := secondary.NewGistPusher()
		if err := pusher.Push(gistID, token, gistFile, content); err != nil {
			return fmt.Errorf("failed to push gist: %w", err)
		}

		logger.Info("Pushed to gist", "gist_id", gistID, "source", input)
		return nil
	},
}

func init() {
	pullCmd.Flags().String("gist-id", "", "GitHub Gist ID")
	pullCmd.Flags().String("gist-file", "", "Filename inside the Gist (uses first file if not specified)")
	pullCmd.Flags().String("github-token", "", "GitHub Personal Access Token (or via GITHUB_TOKEN env)")
	pullCmd.Flags().String("output", defaultLocalConfig, "Local file to write")
	pullCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")

	pushCmd.Flags().String("gist-id", "", "GitHub Gist ID")
	pushCmd.Flags().String("gist-file", "", "Filename inside the Gist to update (uses first file if not specified)")
	pushCmd.Flags().String("github-token", "", "GitHub Personal Access Token (or via GITHUB_TOKEN env)")
	pushCmd.Flags().String("input", defaultLocalConfig, "Local file to upload")
	pushCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")
}

func newLogger(verbose bool) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	if verbose {
		opts.Level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

func cacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".coati", "cache", "gists")
}

func resolveCredentials(token, gistID string, logger *slog.Logger) (string, string) {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" || gistID == "" {
		cfg, err := services.NewConfigManager(logger, nil).LoadConfig()
		if err == nil {
			if token == "" {
				token = cfg.GitHubToken
			}
			if gistID == "" {
				gistID = cfg.GistID
			}
		}
	}
	return token, gistID
}
