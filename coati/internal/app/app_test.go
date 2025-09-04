package app

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplication_CreateBackup(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "hosts")
	originalContent := []byte("127.0.0.1 original-host\n")

	// Pre-create the target file
	err := os.WriteFile(targetFile, originalContent, 0644)
	require.NoError(t, err)

	// Create application instance
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := Config{
		Backup: true,
	}
	app, err := New(logger, cfg)
	require.NoError(t, err)

	// Run backup
	err = app.createBackup(targetFile)
	require.NoError(t, err)

	// Check if backup exists and has the correct content
	backupFile := targetFile + ".bak"
	require.FileExists(t, backupFile)

	backupContent, err := os.ReadFile(backupFile)
	require.NoError(t, err)
	assert.Equal(t, originalContent, backupContent)

	// Ensure permissions match (0644)
	info, err := os.Stat(backupFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
}

func TestApplication_CreateBackup_NotExist(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "hosts-nonexistent")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := Config{
		Backup: true,
	}
	app, err := New(logger, cfg)
	require.NoError(t, err)

	// Should not return error and should not create backup
	err = app.createBackup(targetFile)
	require.NoError(t, err)

	backupFile := targetFile + ".bak"
	assert.NoFileExists(t, backupFile)
}
