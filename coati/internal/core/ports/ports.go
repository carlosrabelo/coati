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

type HostsGenerator interface {
	GenerateHosts() ([]byte, error)
}

type SSHGenerator interface {
	GenerateSSHConfig() ([]byte, error)
}

type GistFetcher interface {
	Fetch(gistID, token, gistFile string) ([]byte, error)
}

type GistPusher interface {
	Push(gistID, token, gistFile string, content []byte) error
}

type HostsParser interface {
	ParseHosts(content []byte) ([]domain.HostEntry, error)
}

type SSHParser interface {
	ParseSSHConfig(content []byte) ([]domain.SSHConfig, error)
}
