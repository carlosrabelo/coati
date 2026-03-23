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

func TestHostFileParser_ParseHosts(t *testing.T) {
	parser := NewHostFileParser()
	input := []byte(`
# Comment
127.0.0.1 localhost
192.168.1.10 web01 # Web Server
10.0.0.5	db01	db-alias
`)

	entries, err := parser.ParseHosts(input)
	require.NoError(t, err)
	require.Len(t, entries, 3)

	assert.Equal(t, "127.0.0.1", entries[0].IP)
	assert.Equal(t, "localhost", entries[0].Hostname)

	assert.Equal(t, "192.168.1.10", entries[1].IP)
	assert.Equal(t, "web01", entries[1].Hostname)
	assert.Equal(t, "Web Server", entries[1].Comment)

	assert.Equal(t, "10.0.0.5", entries[2].IP)
	assert.Equal(t, "db01", entries[2].Hostname)
	assert.Contains(t, entries[2].Aliases, "db-alias")
}

func TestSSHFileParser_ParseSSHConfig(t *testing.T) {
	parser := NewSSHFileParser()
	input := []byte(`
Host web
	HostName 192.168.1.10
	User admin
	Port 2222
	IdentityFile ~/.ssh/id_web

Host db
	HostName 10.0.0.5
	User root
	StrictHostKeyChecking no
`)

	configs, err := parser.ParseSSHConfig(input)
	require.NoError(t, err)
	require.Len(t, configs, 2)

	// Check web
	assert.Equal(t, "web", configs[0].Host)
	assert.Equal(t, "192.168.1.10", configs[0].HostName)
	assert.Equal(t, "admin", configs[0].User)
	assert.Equal(t, 2222, configs[0].Port)
	assert.Equal(t, "~/.ssh/id_web", configs[0].IdentityFile)

	// Check db
	assert.Equal(t, "db", configs[1].Host)
	assert.Equal(t, "10.0.0.5", configs[1].HostName)
	assert.Equal(t, "root", configs[1].User)
	assert.Equal(t, "no", configs[1].Options["StrictHostKeyChecking"])
}
