package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationMetrics_Report(t *testing.T) {
	m := NewMetrics()
	require.False(t, m.StartTime.IsZero())

	m.FetchDuration = 1 * time.Millisecond
	m.ParseDuration = 2 * time.Millisecond
	m.HostsGenDuration = 3 * time.Millisecond
	m.SSHGenDuration = 4 * time.Millisecond

	report := m.Report()
	assert.True(t, m.TotalDuration >= 0)
	assert.True(t, strings.HasSuffix(report, "\n"))

	lines := strings.Split(strings.TrimSuffix(report, "\n"), "\n")
	require.Len(t, lines, 6)
	assert.Equal(t, "Performance Metrics:", lines[0])
	assert.Equal(t, "- Fetch Config: 1ms", lines[1])
	assert.Equal(t, "- Parse Config: 2ms", lines[2])
	assert.Equal(t, "- Generate Hosts: 3ms", lines[3])
	assert.Equal(t, "- Generate SSH: 4ms", lines[4])
	assert.True(t, strings.HasPrefix(lines[5], "- Total Execution: "))
}
