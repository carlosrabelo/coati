package domain

import (
	"fmt"
	"time"
)

type ApplicationMetrics struct {
	StartTime        time.Time
	FetchDuration    time.Duration
	ParseDuration    time.Duration
	HostsGenDuration time.Duration
	SSHGenDuration   time.Duration
	TotalDuration    time.Duration
}

func NewMetrics() *ApplicationMetrics {
	return &ApplicationMetrics{
		StartTime: time.Now(),
	}
}

func (m *ApplicationMetrics) Report() string {
	m.TotalDuration = time.Since(m.StartTime)
	return fmt.Sprintf(
		"Performance Metrics:\n"+
			"- Fetch Config: %v\n"+
			"- Parse Config: %v\n"+
			"- Generate Hosts: %v\n"+
			"- Generate SSH: %v\n"+
			"- Total Execution: %v\n",
		m.FetchDuration,
		m.ParseDuration,
		m.HostsGenDuration,
		m.SSHGenDuration,
		m.TotalDuration,
	)
}
