package services

import (
	"errors"
	"strings"
	"testing"

	"coati/coati/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFileReaderForHosts struct {
	mock.Mock
}

func (m *MockFileReaderForHosts) ReadFile(path string) ([]byte, error) {
	args := m.Called(path)
	return args.Get(0).([]byte), args.Error(1)
}


func TestHostsGenerator_GenerateHosts_Valid(t *testing.T) {
	// Setup
	config := domain.GlobalConfig{
		Hosts: []domain.HostConfig{
			{IP: "192.168.1.10", Hostname: "test-host", Comment: "Test Comment"},
			{IP: "10.0.0.1", Hostname: "db-server", Aliases: []string{"db"}},
		},
	}

	// Mock behavior
	mockReader := new(MockFileReaderForHosts)
	mockReader.On("ReadFile", "/etc/hostname").Return([]byte("my-computer"), nil)

	generator := NewHostsGenerator(config, mockReader, "")

	result, err := generator.GenerateHosts()
	assert.NoError(t, err)
	output := string(result)

	assert.Regexp(t, `127\.0\.0\.1\s+localhost`, output)
	assert.Contains(t, output, "my-computer")
	assert.Regexp(t, `192\.168\.1\.10\s+test-host\s+# Test Comment`, output)
	assert.Regexp(t, `10\.0\.0\.1\s+db-server\s+db`, output)

	mockReader.AssertExpectations(t)
}

func TestHostsGenerator_GenerateHosts_SortingAndAliases(t *testing.T) {
	config := domain.GlobalConfig{
		Hosts: []domain.HostConfig{
			{IP: "10.0.0.20", Hostname: "z-host"},
			{IP: "10.0.0.10", Hostname: "a-host"},
		},
		CNAMEs: []domain.CNAMEConfig{
			{Target: "a-host", Aliases: []string{"web-alias"}},
		},
		PTRs: []domain.PTRConfig{
			{IP: "10.0.0.10", Domain: "ptr.example.com"},
		},
	}

	mockReader := new(MockFileReaderForHosts)
	mockReader.On("ReadFile", "/etc/hostname").Return([]byte("my-computer"), nil)

	generator := NewHostsGenerator(config, mockReader, "")

	result, err := generator.GenerateHosts()
	assert.NoError(t, err)
	output := string(result)

	// Verify sorting (a-host 10.0.0.10 should come before z-host 10.0.0.20 based on IP numerical order)
	// Actually logic sorts by IP first using bytes.Compare for IPs.
	// 10.0.0.10 vs 10.0.0.20
	// 10.0.0.10 is numerically smaller.
	assert.Regexp(t, `10\.0\.0\.10\s+a-host\s+ptr\.example\.com\s+web-alias`, output)
	assert.Regexp(t, `10\.0\.0\.20\s+z-host`, output)

	idx1 := strings.Index(output, "10.0.0.10")
	idx2 := strings.Index(output, "10.0.0.20")
	assert.True(t, idx1 < idx2, "Expected 10.0.0.10 to appear before 10.0.0.20")
}

func TestHostsGenerator_GenerateHosts_WithTemplate(t *testing.T) {
	config := domain.GlobalConfig{
		Hosts: []domain.HostConfig{
			{IP: "1.1.1.1", Hostname: "custom-host"},
		},
	}

	mockReader := new(MockFileReaderForHosts)

	mockReader.On("ReadFile", "/etc/hostname").Return([]byte("my-computer"), nil)

	template := `
127.0.0.1 localhost
# Custom Template
<hostname> 127.0.1.1
`
	generator := NewHostsGenerator(config, mockReader, template)

	result, err := generator.GenerateHosts()
	assert.NoError(t, err)
	output := string(result)

	assert.Contains(t, output, "# Custom Template")
	assert.Regexp(t, `my-computer\s+127\.0\.1\.1`, output)
	assert.Regexp(t, `1\.1\.1\.1\s+custom-host`, output)
}

func TestHostsGenerator_GenerateHosts_SimplifiedMode(t *testing.T) {
	config := domain.GlobalConfig{
		SimplifiedModeHosts: []string{"my-computer"},
		Hosts: []domain.HostConfig{
			{IP: "10.0.0.5", Hostname: "simple-host", DNSName: "dns.name", Aliases: []string{"alias"}},
		},
		PTRs: []domain.PTRConfig{
			{IP: "10.0.0.5", Domain: "ptr.domain"},
		},
	}

	mockReader := new(MockFileReaderForHosts)

	mockReader.On("ReadFile", "/etc/hostname").Return([]byte("my-computer"), nil)

	generator := NewHostsGenerator(config, mockReader, "")

	result, err := generator.GenerateHosts()
	assert.NoError(t, err)
	output := string(result)

	// In simplified mode:
	// 1. PTR/DNSName are NOT added as first alias.
	// 2. CNAMEs are NOT added.
	// 3. Explicit aliases ARE added.

	// Expect: 10.0.0.5 simple-host alias
	// NOT expect: ptr.domain or dns.name
	assert.Regexp(t, `10\.0\.0\.5\s+simple-host\s+alias`, output)
	assert.NotContains(t, output, "ptr.domain")
	assert.NotContains(t, output, "dns.name")
}

func TestHostsGenerator_GenerateHosts_CNAMEviaDNSName(t *testing.T) {
	config := domain.GlobalConfig{
		Hosts: []domain.HostConfig{
			{IP: "10.0.0.10", Hostname: "srv", DNSName: "srv.example.com"},
		},
		CNAMEs: []domain.CNAMEConfig{
			{Target: "srv.example.com", Aliases: []string{"www.example.com"}},
		},
	}

	mockReader := new(MockFileReaderForHosts)
	mockReader.On("ReadFile", "/etc/hostname").Return([]byte("my-computer"), nil)

	generator := NewHostsGenerator(config, mockReader, "")
	result, err := generator.GenerateHosts()
	assert.NoError(t, err)
	output := string(result)

	// CNAME matched via DNSName: www.example.com should appear as alias
	assert.Contains(t, output, "www.example.com")
	mockReader.AssertExpectations(t)
}

func TestHostsGenerator_GenerateHosts_ReadFileError(t *testing.T) {
	config := domain.GlobalConfig{
		Hosts: []domain.HostConfig{
			{IP: "10.0.0.1", Hostname: "some-host"},
		},
	}

	mockReader := new(MockFileReaderForHosts)
	mockReader.On("ReadFile", "/etc/hostname").Return([]byte(nil), errors.New("read error"))

	generator := NewHostsGenerator(config, mockReader, "")
	result, err := generator.GenerateHosts()
	assert.NoError(t, err)
	output := string(result)

	// Falls back to "localhost" when hostname file cannot be read
	assert.Contains(t, output, "localhost")
	assert.Contains(t, output, "some-host")
	mockReader.AssertExpectations(t)
}

func TestHostsGenerator_GenerateHosts_EmptyHosts(t *testing.T) {
	config := domain.GlobalConfig{}

	mockReader := new(MockFileReaderForHosts)
	mockReader.On("ReadFile", "/etc/hostname").Return([]byte("my-computer"), nil)

	generator := NewHostsGenerator(config, mockReader, "")
	result, err := generator.GenerateHosts()
	assert.NoError(t, err)
	output := string(result)

	// Should still have the template block header and no host entries from config
	assert.Contains(t, output, "# === Source: Template (Default) ===")
	assert.NotContains(t, output, "# === Source: Configuration ===")
	mockReader.AssertExpectations(t)
}

func TestHostsGenerator_GenerateHosts_CNAME_CaseInsensitive(t *testing.T) {
	config := domain.GlobalConfig{
		Hosts: []domain.HostConfig{
			{IP: "10.0.0.10", Hostname: "srv", DNSName: "srv.example.com"},
		},
		CNAMEs: []domain.CNAMEConfig{
			{Target: "SRV.Example.COM", Aliases: []string{"www.example.com"}},
		},
	}

	mockReader := new(MockFileReaderForHosts)
	mockReader.On("ReadFile", "/etc/hostname").Return([]byte("my-computer"), nil)

	generator := NewHostsGenerator(config, mockReader, "")
	result, err := generator.GenerateHosts()
	assert.NoError(t, err)
	output := string(result)

	// CNAME with mixed-case target should match lowercase DNSName
	assert.Contains(t, output, "www.example.com")
	mockReader.AssertExpectations(t)
}

func TestHostsGenerator_GenerateHosts_ByIP(t *testing.T) {
	config := domain.GlobalConfig{
		Hosts: []domain.HostConfig{
			{IP: "10.0.0.2", Hostname: "a-host"}, // Should come second (IP > 10.0.0.1), even though 'a' < 'z'
			{IP: "10.0.0.1", Hostname: "z-host"}, // Should come first
		},
	}

	mockReader := new(MockFileReaderForHosts)
	mockReader.On("ReadFile", "/etc/hostname").Return([]byte("my-computer"), nil)

	generator := NewHostsGenerator(config, mockReader, "")

	result, err := generator.GenerateHosts()
	assert.NoError(t, err)
	output := string(result)

	// Verify headers do NOT exist
	assert.NotContains(t, output, "# === Production ===")
	assert.NotContains(t, output, "# === Development ===")
	assert.NotContains(t, output, "# === Misc ===")

	// Verify Source Headers exist
	assert.Contains(t, output, "# === Source: Template (Default) ===")
	assert.Contains(t, output, "# === Source: Configuration ===")

	// Verify Sort Order by IP:
	// z-host (10.0.0.1) should come before a-host (10.0.0.2)

	idxZHost := strings.Index(output, "z-host")
	idxAHost := strings.Index(output, "a-host")

	assert.True(t, idxZHost < idxAHost, "z-host (10.0.0.1) should be before a-host (10.0.0.2) because of IP sorting")
}
