package secondary

// Package secondary provides filesystem and HTTP adapter implementations.

import (
	"os"
	"path/filepath"
)

type FSAdapter struct{}

func NewFSAdapter() *FSAdapter {
	return &FSAdapter{}
}

func (a *FSAdapter) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (a *FSAdapter) WriteFile(path string, content []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, content, perm)
}

