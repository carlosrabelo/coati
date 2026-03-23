package services

import (
	"strings"
	"testing"

	"coati/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestSSHGenerator_GenerateSSHConfig_Valid(t *testing.T) {
	// Setup
	config := domain.GlobalConfig{
		Hosts: []domain.HostConfig{
			{IP: "192.168.1.50", Hostname: "app-server", User: "deploy", Port: 2222},
		},
		Defaults: domain.SSHDefaults{
			User: "admin",
		},
	}

	generator := NewSSHGenerator(config)

	// Execute
	result, err := generator.GenerateSSHConfig()

	// Assert
	assert.NoError(t, err)
	output := string(result)

	assert.Contains(t, output, "Host app-server")
	assert.Contains(t, output, "HostName 192.168.1.50")
	assert.Contains(t, output, "User deploy")
	assert.Contains(t, output, "Port 2222")
	assert.Contains(t, output, "Host *")
	assert.Contains(t, output, "User admin")

}

func TestSSHGenerator_GenerateSSHConfig_Alphabetical(t *testing.T) {
	config := domain.GlobalConfig{
		Hosts: []domain.HostConfig{
			{IP: "1.1.1.1", Hostname: "srv-b"},
			{IP: "2.2.2.2", Hostname: "srv-a"},
		},
	}

	generator := NewSSHGenerator(config)
	result, err := generator.GenerateSSHConfig()
	assert.NoError(t, err)
	output := string(result)

	// Verify NO headers
	assert.NotContains(t, output, "# === GroupA ===")
	assert.NotContains(t, output, "# === GroupB ===")

	idxSrvA := strings.Index(output, "srv-a")
	idxSrvB := strings.Index(output, "srv-b")
	// srv-b (1.1.1.1) < srv-a (2.2.2.2)
	assert.True(t, idxSrvB < idxSrvA, "srv-b should appear before srv-a because 1.1.1.1 < 2.2.2.2")
}

func TestSSHGenerator_GenerateSSHConfig_EmptyHosts(t *testing.T) {
	config := domain.GlobalConfig{
		Defaults: domain.SSHDefaults{
			User: "admin",
		},
	}

	generator := NewSSHGenerator(config)
	result, err := generator.GenerateSSHConfig()
	assert.NoError(t, err)
	output := string(result)

	// No hosts, but defaults block should still appear
	assert.Contains(t, output, "Host *")
	assert.Contains(t, output, "User admin")
	assert.NotContains(t, output, "HostName")
}

func TestSSHGenerator_GenerateSSHConfig_NoDefaults(t *testing.T) {
	config := domain.GlobalConfig{
		Hosts: []domain.HostConfig{
			{IP: "10.0.0.1", Hostname: "srv"},
		},
	}

	generator := NewSSHGenerator(config)
	result, err := generator.GenerateSSHConfig()
	assert.NoError(t, err)
	output := string(result)

	assert.Contains(t, output, "Host srv")
	assert.NotContains(t, output, "Host *")
}

func TestSSHGenerator_GenerateSSHConfig_SecondarySort(t *testing.T) {
	config := domain.GlobalConfig{
		Hosts: []domain.HostConfig{
			{IP: "1.1.1.1", Hostname: "srv-y"},
			{IP: "1.1.1.1", Hostname: "srv-x"},
		},
	}

	generator := NewSSHGenerator(config)
	result, err := generator.GenerateSSHConfig()
	assert.NoError(t, err)
	output := string(result)

	idxSrvX := strings.Index(output, "srv-x")
	idxSrvY := strings.Index(output, "srv-y")

	// Same IP, so compare Hostname: srv-x < srv-y
	assert.True(t, idxSrvX < idxSrvY, "srv-x should appear before srv-y because srv-x < srv-y")
}
