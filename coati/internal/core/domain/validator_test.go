package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateIP(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		ip    string
		valid bool
	}{
		{"192.168.1.1", true},
		{"127.0.0.1", true},
		{"::1", true},
		{"fe80::1", true},
		{"256.0.0.1", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		err := v.ValidateIP(tt.ip)
		if tt.valid {
			assert.NoError(t, err, "IP: %s", tt.ip)
		} else {
			assert.Error(t, err, "IP: %s", tt.ip)
		}
	}
}

func TestValidator_ValidateHostname(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		host  string
		valid bool
	}{
		{"localhost", true},
		{"example.com", true},
		{"my-host", true},
		{"sub.domain.co.uk", true},
		{"-start", false},
		{"end-", false},
		{"in valid", false}, // space
		{"", false},
	}

	for _, tt := range tests {
		err := v.ValidateHostname(tt.host)
		if tt.valid {
			assert.NoError(t, err, "Host: %s", tt.host)
		} else {
			assert.Error(t, err, "Host: %s", tt.host)
		}
	}
}

func TestValidator_ValidateHostEntry(t *testing.T) {
	v := NewValidator()

	entry := HostConfig{
		IP:       "192.168.1.50",
		Hostname: "valid-host",
		Aliases:  []string{"alias1", "alias2"},
	}
	assert.NoError(t, v.ValidateHostEntry(entry))

	badIP := HostConfig{
		IP:       "bad-ip",
		Hostname: "valid-host",
	}
	assert.Error(t, v.ValidateHostEntry(badIP))

	badAlias := HostConfig{
		IP:       "192.168.1.50",
		Hostname: "valid-host",
		Aliases:  []string{"bad alias"},
	}
	assert.Error(t, v.ValidateHostEntry(badAlias))
}

func TestGlobalConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  GlobalConfig
		wantErr bool
	}{
		{
			name: "valid IP entry",
			config: GlobalConfig{
				Hosts: []HostConfig{{IP: "192.168.1.1", Hostname: "web"}},
			},
		},
		{
			name: "SSH-only entry with hostname in ip field",
			config: GlobalConfig{
				Hosts: []HostConfig{{IP: "github.com", Hostname: "gh"}},
			},
		},
		{
			name: "invalid ip field",
			config: GlobalConfig{
				Hosts: []HostConfig{{IP: "not a valid ip or hostname!", Hostname: "ok"}},
			},
			wantErr: true,
		},
		{
			name: "invalid hostname",
			config: GlobalConfig{
				Hosts: []HostConfig{{IP: "10.0.0.1", Hostname: "-bad-"}},
			},
			wantErr: true,
		},
		{
			name: "invalid alias",
			config: GlobalConfig{
				Hosts: []HostConfig{{IP: "10.0.0.1", Hostname: "ok", Aliases: []string{"bad alias"}}},
			},
			wantErr: true,
		},
		{
			name: "invalid default port",
			config: GlobalConfig{
				Defaults: SSHDefaults{Port: 99999},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
