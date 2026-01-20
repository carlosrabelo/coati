package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
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

func splitCommand(line string) ([]string, error) {
	var args []string
	var current strings.Builder
	inDoubleQuotes := false
	inSingleQuotes := false
	escaped := false

	for i := 0; i < len(line); i++ {
		r := line[i]

		if escaped {
			current.WriteByte(r)
			escaped = false
			continue
		}

		if r == '\\' && !inSingleQuotes {
			escaped = true
			continue
		}

		if r == '"' && !inSingleQuotes {
			inDoubleQuotes = !inDoubleQuotes
			continue
		}

		if r == '\'' && !inDoubleQuotes {
			inSingleQuotes = !inSingleQuotes
			continue
		}

		if (r == ' ' || r == '\t' || r == '\n' || r == '\r') && !inDoubleQuotes && !inSingleQuotes {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteByte(r)
	}

	if inDoubleQuotes || inSingleQuotes {
		return nil, fmt.Errorf("unclosed quotes in command")
	}
	if escaped {
		return nil, fmt.Errorf("trailing backslash in command")
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args, nil
}

func (app *Application) validateHookCommand(hook string) error {
	parts, err := splitCommand(hook)
	if err != nil {
		return fmt.Errorf("failed to parse hook: %w", err)
	}
	if len(parts) == 0 {
		return nil
	}

	if strings.ContainsAny(parts[0], "/\\") && !app.cfg.AllowUnsafeHooks {
		return fmt.Errorf("hook command must be a command name only, not a path: %s", parts[0])
	}

	commandName := parts[0]
	if !app.cfg.AllowUnsafeHooks {
		isAllowed := allowedHookCommands[commandName]
		if !isAllowed {
			// Check custom allowed hooks
			for _, allowed := range app.cfg.AllowedHooks {
				if allowed == commandName {
					isAllowed = true
					break
				}
			}
		}

		if !isAllowed {
			return fmt.Errorf("command not allowed: %s (allowed commands: %v)", commandName, app.getAllowedCommandNames())
		}
	}

	for _, part := range parts[1:] {
		if strings.Contains(part, ";") || strings.Contains(part, "&") || strings.Contains(part, "|") {
			return fmt.Errorf("command contains forbidden characters: %s", part)
		}
	}

	return nil
}

func (app *Application) getAllowedCommandNames() []string {
	names := make([]string, 0, len(allowedHookCommands)+len(app.cfg.AllowedHooks))
	for cmd := range allowedHookCommands {
		names = append(names, cmd)
	}
	for _, cmd := range app.cfg.AllowedHooks {
		names = append(names, cmd)
	}
	return names
}

// runHooks validates all hooks first, then executes them in order.
// A validation failure aborts before any hook runs.
func (app *Application) runHooks(hooks []string) error {
	for _, hook := range hooks {
		if err := app.validateHookCommand(hook); err != nil {
			app.logger.Error("Hook validation failed", "command", hook, "error", err)
			return fmt.Errorf("hook validation failed: %w", err)
		}
	}
	for _, hook := range hooks {
		parts, _ := splitCommand(hook)
		if len(parts) == 0 {
			continue
		}
		app.logger.Info("Executing hook", "command", hook)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			app.logger.Warn("Hook execution failed", "command", hook, "error", err)
		}
	}
	return nil
}
