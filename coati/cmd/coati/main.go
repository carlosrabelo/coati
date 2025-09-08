package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Execute runs the root Cobra command.
func Execute() error {
	rootCmd := &cobra.Command{
		Use:   "coati",
		Short: "Manage /etc/hosts and ~/.ssh/config from a YAML definition",
	}
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.AddCommand(newProcessCmd())
	return rootCmd.Execute()
}
