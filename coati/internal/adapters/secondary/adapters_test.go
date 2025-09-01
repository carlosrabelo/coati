package secondary

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFSAdapter_ReadWriteFile(t *testing.T) {
	adapter := NewFSAdapter()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := []byte("hello world")

	// Test Write
	err := adapter.WriteFile(filePath, content, 0644)
	require.NoError(t, err)

	// Test Read
	readContent, err := adapter.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, content, readContent)
}
