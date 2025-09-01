package ports

import (
	"os"
)

type FileReader interface {
	ReadFile(path string) ([]byte, error)
}

type FileWriter interface {
	WriteFile(path string, content []byte, perm os.FileMode) error
}
