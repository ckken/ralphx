package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func WorkersDir(p Paths) string { return filepath.Join(p.Root, "workers") }
func ResultsDir(p Paths) string { return filepath.Join(p.Root, "results") }

func EnsureParallelDirs(p Paths) error {
	for _, dir := range []string{WorkersDir(p), ResultsDir(p)} {
		if err := os.MkdirAll(dir, defaultDirMode); err != nil {
			return err
		}
	}
	return nil
}

func WriteJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), defaultDirMode); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, defaultFileMode)
}
