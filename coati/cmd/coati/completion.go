package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:                   "completion [bash|zsh|fish|powershell]",
	Short:                 "Install shell completion",
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		switch args[0] {
		case "bash":
			dest := filepath.Join(home, ".local", "share", "bash-completion", "completions", "coati")
			if err := writeCompletion(dest, func() error {
				return cmd.Root().GenBashCompletionFile(dest)
			}); err != nil {
				return err
			}
			fmt.Printf("Installed: %s\nReload your shell or run: source %s\n", dest, dest)

		case "zsh":
			dest := filepath.Join(home, ".zfunc", "_coati")
			if err := writeCompletion(dest, func() error {
				f, err := os.Create(dest)
				if err != nil {
					return err
				}
				defer f.Close()
				return cmd.Root().GenZshCompletion(f)
			}); err != nil {
				return err
			}
			fmt.Printf("Installed: %s\n", dest)
			fmt.Println("If not already set, add to ~/.zshrc:")
			fmt.Println("  fpath=(~/.zfunc $fpath)")
			fmt.Println("  autoload -U compinit && compinit")

		case "fish":
			dest := filepath.Join(home, ".config", "fish", "completions", "coati.fish")
			if err := writeCompletion(dest, func() error {
				f, err := os.Create(dest)
				if err != nil {
					return err
				}
				defer f.Close()
				return cmd.Root().GenFishCompletion(f, true)
			}); err != nil {
				return err
			}
			fmt.Printf("Installed: %s\nCompletions are active in new fish sessions.\n", dest)

		case "powershell":
			fmt.Println("# Add the following to your PowerShell profile ($PROFILE):")
			fmt.Println()
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
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
