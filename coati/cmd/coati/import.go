package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"coati/coati/internal/adapters/secondary"
	"coati/coati/internal/app"
	"coati/coati/internal/core/services"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import existing /etc/hosts and ~/.ssh/config into a YAML configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		hostsPath, _ := cmd.Flags().GetString("hosts-file")
		sshPath, _ := cmd.Flags().GetString("ssh-file")
		outputPath, _ := cmd.Flags().GetString("output")
		verbose, _ := cmd.Flags().GetBool("verbose")

		logger := newLogger(verbose)

		hostsPath = expandPath(hostsPath)
		sshPath = expandPath(sshPath)

		var hostsContent []byte
		var err error
		if hostsPath != "" {
			if _, err := os.Stat(hostsPath); err == nil {
				logger.Debug("Reading hosts file", "path", hostsPath)
				hostsContent, err = os.ReadFile(hostsPath)
				if err != nil {
					return fmt.Errorf("failed to read hosts file %s: %w", hostsPath, err)
				}
			} else {
				logger.Warn("Hosts file not found, skipping", "path", hostsPath)
			}
		}

		var sshContent []byte
		if sshPath != "" {
			if _, err := os.Stat(sshPath); err == nil {
				logger.Debug("Reading SSH config file", "path", sshPath)
				sshContent, err = os.ReadFile(sshPath)
				if err != nil {
					return fmt.Errorf("failed to read SSH config file %s: %w", sshPath, err)
				}
			} else {
				logger.Warn("SSH config file not found, skipping", "path", sshPath)
			}
		}

		if len(hostsContent) == 0 && len(sshContent) == 0 {
			return fmt.Errorf("both hosts and SSH config files were not found or empty")
		}

		hostsParser := secondary.NewHostFileParser()
		sshParser := secondary.NewSSHFileParser()
		importer := services.NewImporter(hostsParser, sshParser)

		logger.Info("Merging configurations...")
		config, err := importer.Import(hostsContent, sshContent)
		if err != nil {
			return fmt.Errorf("failed to import: %w", err)
		}

		yamlData, err := yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config to YAML: %w", err)
		}

		if outputPath == "-" {
			fmt.Println(string(yamlData))
		} else {
			logger.Info("Writing configuration", "path", outputPath)
			if err := app.WriteAtomicPublic(outputPath, yamlData, 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			logger.Info("Import completed successfully")
		}

		return nil
	},
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func init() {
	importCmd.Flags().String("hosts-file", "/etc/hosts", "Path to hosts file")
	importCmd.Flags().String("ssh-file", "~/.ssh/config", "Path to SSH config file")
	importCmd.Flags().String("output", "hosts.yaml", "Path to save the generated YAML configuration (use '-' for stdout)")
	importCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")
}
