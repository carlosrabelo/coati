package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var allowedHookCommands = map[string]bool{
	"systemctl": true,
	"service":   true,
	"docker":    true,
	"kubectl":   true,
	"nginx":     true,
	"apache2":   true,
	"httpd":     true,
}

func validateHookCommand(hook string) error {
	parts := strings.Fields(hook)
	if len(parts) == 0 {
		return nil
	}

	commandName := filepath.Base(parts[0])
	if !allowedHookCommands[commandName] {
		return fmt.Errorf("command not allowed: %s (allowed commands: %v)", commandName, getAllowedCommandNames())
	}

	for _, part := range parts[1:] {
		if strings.Contains(part, ";") || strings.Contains(part, "&") || strings.Contains(part, "|") {
			return fmt.Errorf("command contains forbidden characters: %s", part)
		}
	}

	return nil
}

func getAllowedCommandNames() []string {
	names := make([]string, 0, len(allowedHookCommands))
	for cmd := range allowedHookCommands {
		names = append(names, cmd)
	}
	return names
}

// runHooks validates all hooks first, then executes them in order.
// A validation failure aborts before any hook runs.
func (app *Application) runHooks(hooks []string) error {
	for _, hook := range hooks {
		if err := validateHookCommand(hook); err != nil {
			app.logger.Error("Hook validation failed", "command", hook, "error", err)
			return fmt.Errorf("hook validation failed: %w", err)
		}
	}
	for _, hook := range hooks {
		parts := strings.Fields(hook)
		if len(parts) == 0 {
			continue
		}
		app.logger.Info("Executing hook", "command", hook)
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			app.logger.Warn("Hook execution failed", "command", hook, "error", err)
		}
	}
	return nil
}
