package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateHookCommand_Valid(t *testing.T) {
	app := &Application{}

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
			err := app.validateHookCommand(tt.hook)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateHookCommand_UnsafeAndCustom(t *testing.T) {
	// Test Custom Allowed Command
	appCustom := &Application{
		cfg: Config{
			AllowedHooks: []string{"dnsmasq", "custom-script"},
		},
	}

	err := appCustom.validateHookCommand("dnsmasq reload")
	assert.NoError(t, err)

	err = appCustom.validateHookCommand("custom-script run")
	assert.NoError(t, err)

	err = appCustom.validateHookCommand("rm -rf /")
	assert.Error(t, err)

	// Test Unsafe Hooks
	appUnsafe := &Application{
		cfg: Config{
			AllowUnsafeHooks: true,
		},
	}

	err = appUnsafe.validateHookCommand("rm -rf /")
	assert.NoError(t, err)

	// Injection is still blocked even with unsafe hooks enabled
	err = appUnsafe.validateHookCommand("rm -rf /; touch /hack")
	assert.Error(t, err)
}

func TestGetAllowedCommandNames(t *testing.T) {
	app := &Application{
		cfg: Config{
			AllowedHooks: []string{"custom-cmd"},
		},
	}
	names := app.getAllowedCommandNames()
	assert.NotEmpty(t, names)
	assert.Contains(t, names, "systemctl")
	assert.Contains(t, names, "docker")
	assert.Contains(t, names, "custom-cmd")
}
