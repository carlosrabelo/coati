package ports

import (
	"os"

	"coati/coati/internal/core/domain"
)

type FileReader interface {
	ReadFile(path string) ([]byte, error)
}

type FileWriter interface {
	WriteFile(path string, content []byte, perm os.FileMode) error
}

type GistFetcher interface {
	Fetch(gistID, token, gistFile string) ([]byte, error)
}

type HostsParser interface {
	ParseHosts(content []byte) ([]domain.HostEntry, error)
}

type SSHParser interface {
	ParseSSHConfig(content []byte) ([]domain.SSHConfig, error)
}
