package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateHookCommand_Valid(t *testing.T) {
	tests := []struct {
		name    string
		hook    string
		wantErr bool
	}{
		{
			name:    "valid systemctl command",
			hook:    "systemctl restart nginx",
			wantErr: false,
		},
		{
			name:    "valid docker command",
			hook:    "docker restart container",
			wantErr: false,
		},
		{
			name:    "valid kubectl command",
			hook:    "kubectl apply -f deployment.yaml",
			wantErr: false,
		},
		{
			name:    "empty hook",
			hook:    "",
			wantErr: false,
		},
		{
			name:    "invalid command - not in allowlist",
			hook:    "rm -rf /",
			wantErr: true,
		},
		{
			name:    "forbidden semicolon",
			hook:    "systemctl restart nginx; rm -rf /",
			wantErr: true,
		},
		{
			name:    "forbidden ampersand",
			hook:    "systemctl restart nginx &",
			wantErr: true,
		},
		{
			name:    "forbidden pipe",
			hook:    "systemctl status nginx | grep running",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHookCommand(tt.hook)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetAllowedCommandNames(t *testing.T) {
	names := getAllowedCommandNames()
	assert.NotEmpty(t, names)
	assert.Contains(t, names, "systemctl")
	assert.Contains(t, names, "docker")
}
