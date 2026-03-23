package ports

import "os"

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
	Fetch(gistID, token string) ([]byte, error)
}
