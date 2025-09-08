package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"coati/coati/internal/core/services"
)

var completionCmd = &cobra.Command{
	Use:                   "completion [bash|zsh|fish]",
	Short:                 "Install system-wide shell completion (root only)",
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Long: `Install shell completion system-wide under /usr/local.

Requires root. Choose the shell:

  sudo coati completion bash
  sudo coati completion zsh
  sudo coati completion fish
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !services.RunningAsRoot() {
			return fmt.Errorf("completion requires root; run:\n  sudo coati completion %s", args[0])
		}

		root := cmd.Root()
		switch args[0] {
		case "bash":
			dest := "/usr/local/share/bash-completion/completions/coati"
			if err := writeCompletion(dest, func() error {
				return root.GenBashCompletionFile(dest)
			}); err != nil {
				return err
			}
			fmt.Printf("Installed: %s\n", dest)

		case "zsh":
			dest := "/usr/local/share/zsh/site-functions/_coati"
			if err := writeCompletion(dest, func() error {
				f, err := os.Create(dest)
				if err != nil {
					return err
				}
				defer f.Close()
				return root.GenZshCompletion(f)
			}); err != nil {
				return err
			}
			fmt.Printf("Installed: %s\n", dest)

		case "fish":
			dest := "/usr/local/share/fish/vendor_completions.d/coati.fish"
			if err := writeCompletion(dest, func() error {
				f, err := os.Create(dest)
				if err != nil {
					return err
				}
				defer f.Close()
				return root.GenFishCompletion(f, true)
			}); err != nil {
				return err
			}
			fmt.Printf("Installed: %s\n", dest)
		}

		return nil
	},
}

func writeCompletion(dest string, generate func() error) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return generate()
}
