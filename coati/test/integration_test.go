package test

import (
	"os"
	"path/filepath"
	"testing"

	"coati/coati/internal/adapters/secondary"
	"coati/coati/internal/core/domain"
	"coati/coati/internal/core/services"
	"coati/coati/internal/templates"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Simple integration test running strict checking over generated files
// This assumes we can compile and run main or reuse internal packages in a test context.
// Since we are in the same module, we can import packages.
// Ideally, we'd run the binary, but importing packages allows easier checking without shelling out.

// NOTE: We need to export Application or expose a Run method that we can call with custom args/dependencies
// OR we can just use the services directly here to simulate the flow.
// 'main' package is not importable. So we will test the flow using services integration.

func TestIntegration_EndToEndGeneration(t *testing.T) {
	// Setup temporary files
	tmpDir := t.TempDir()
	currentHostsPath := filepath.Join(tmpDir, "current_hosts")
	currentSSHPath := filepath.Join(tmpDir, "current_ssh_config")

	// Create fake current files
	err := os.WriteFile(currentHostsPath, []byte("127.0.0.1 localhost\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(currentSSHPath, []byte("Host existing\n  HostName 1.2.3.4\n"), 0644)
	require.NoError(t, err)

	// Config Input
	config := domain.GlobalConfig{
		Hosts: []domain.HostConfig{
			{IP: "192.168.100.10", Hostname: "integration-host", User: "root", Port: 2222},
		},
		Defaults: domain.SSHDefaults{
			User: "ubuntu",
		},
	}

	// Initialize Real Adapters
	fsAdapter := secondary.NewFSAdapter()

	// Initialize Generators
	// Initialize Generators
	hostsGen := services.NewHostsGenerator(config, fsAdapter, templates.HostsTemplate)
	sshGen := services.NewSSHGenerator(config)

	// Run Generation
	hostsContent, err := hostsGen.GenerateHosts()
	require.NoError(t, err)

	sshContent, err := sshGen.GenerateSSHConfig()
	require.NoError(t, err)

	// Verify Output
	hostsStr := string(hostsContent)
	sshStr := string(sshContent)

	// Hosts Verification
	assert.Contains(t, hostsStr, "# === Source: Template")
	assert.Contains(t, hostsStr, "# === Source: Configuration")
	assert.Regexp(t, `127\.0\.0\.1\s+localhost`, hostsStr) // From template (defaults)
	assert.Contains(t, hostsStr, "integration-host")
	assert.Contains(t, hostsStr, "192.168.100.10")

	// SSH Verification
	assert.Contains(t, sshStr, "# === Source: Configuration ===")
	assert.Regexp(t, `Host\s+integration-host`, sshStr)
	assert.Regexp(t, `HostName\s+192\.168\.100\.10`, sshStr)
	assert.Regexp(t, `User\s+root`, sshStr)
	assert.Regexp(t, `Port\s+2222`, sshStr)

	// Default SSH verification
	assert.Contains(t, sshStr, "# === Source: Defaults ===")
	assert.Regexp(t, `Host\s+\*`, sshStr)
	assert.Regexp(t, `User\s+ubuntu`, sshStr)

	// Existing SSH verification - SHOULD NOT EXIST ANYMORE
	assert.NotContains(t, sshStr, "Host existing")
}
