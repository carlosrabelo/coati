package services

import (
	"testing"

	"coati/coati/internal/adapters/secondary"
	"coati/coati/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImporter_Import_MergeCorrectly(t *testing.T) {
	hostsContent := []byte(`
# Some comment
127.0.0.1 localhost
192.168.1.10 web01 alias-web # Web server
10.0.0.5 db01
`)

	sshContent := []byte(`
Host *
	User dev
	Port 22

Host web01
	HostName 192.168.1.10
	User admin
	IdentityFile ~/.ssh/id_web

Host db01 db-alias
	HostName 10.0.0.5
	Port 2222

Host only-ssh
	HostName github.com
	User git
`)

	hostsParser := secondary.NewHostFileParser()
	sshParser := secondary.NewSSHFileParser()
	importer := NewImporter(hostsParser, sshParser)

	config, err := importer.Import(hostsContent, sshContent)
	require.NoError(t, err)

	// Verify defaults
	assert.Equal(t, "dev", config.Defaults.User)
	assert.Equal(t, 22, config.Defaults.Port)

	// Verify hosts count
	// We expect 3 hosts: web01, db01, only-ssh
	// localhost should be skipped as it is standard loopback
	require.Len(t, config.Hosts, 3)

	// Find web01
	var webHost *domain.HostConfig
	for i := range config.Hosts {
		if config.Hosts[i].Hostname == "web01" {
			webHost = &config.Hosts[i]
		}
	}
	require.NotNil(t, webHost, "web01 should be imported")
	assert.Equal(t, "192.168.1.10", webHost.IP)
	assert.Contains(t, webHost.Aliases, "alias-web")
	assert.Equal(t, "Web server", webHost.Comment)
	assert.Equal(t, "admin", webHost.User)
	assert.Equal(t, "~/.ssh/id_web", webHost.IdentityFile)

	// Find db01
	var dbHost *domain.HostConfig
	for i := range config.Hosts {
		if config.Hosts[i].Hostname == "db01" {
			dbHost = &config.Hosts[i]
		}
	}
	require.NotNil(t, dbHost, "db01 should be imported")
	assert.Equal(t, "10.0.0.5", dbHost.IP)
	assert.Contains(t, dbHost.Aliases, "db-alias")
	assert.Equal(t, 2222, dbHost.Port)

	// Find only-ssh
	var sshOnlyHost *domain.HostConfig
	for i := range config.Hosts {
		if config.Hosts[i].Hostname == "only-ssh" {
			sshOnlyHost = &config.Hosts[i]
		}
	}
	require.NotNil(t, sshOnlyHost, "only-ssh should be imported")
	assert.Equal(t, "github.com", sshOnlyHost.IP)
	assert.Equal(t, "git", sshOnlyHost.User)
}

func TestImporter_Import_EmptyFiles(t *testing.T) {
	hostsParser := secondary.NewHostFileParser()
	sshParser := secondary.NewSSHFileParser()
	importer := NewImporter(hostsParser, sshParser)

	config, err := importer.Import(nil, nil)
	require.NoError(t, err)
	assert.Empty(t, config.Hosts)
}

func TestImporter_Import_DuplicateIPs(t *testing.T) {
	hostsContent := []byte(`
10.13.250.253 tail1
10.13.250.253 tail2
`)

	sshContent := []byte(`
Host tail1
	HostName 10.13.250.253
	Port 22

Host tail2
	HostName 10.13.250.253
	Port 2222
`)

	hostsParser := secondary.NewHostFileParser()
	sshParser := secondary.NewSSHFileParser()
	importer := NewImporter(hostsParser, sshParser)

	config, err := importer.Import(hostsContent, sshContent)
	require.NoError(t, err)

	require.Len(t, config.Hosts, 2)

	var tail1, tail2 *domain.HostConfig
	for i := range config.Hosts {
		if config.Hosts[i].Hostname == "tail1" {
			tail1 = &config.Hosts[i]
		}
		if config.Hosts[i].Hostname == "tail2" {
			tail2 = &config.Hosts[i]
		}
	}

	require.NotNil(t, tail1)
	require.NotNil(t, tail2)

	assert.Equal(t, "10.13.250.253", tail1.IP)
	assert.Equal(t, 22, tail1.Port)

	assert.Equal(t, "10.13.250.253", tail2.IP)
	assert.Equal(t, 2222, tail2.Port)
}
