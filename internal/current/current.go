package current

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type State struct {
	Path         string
	Version      string
	Binary       string
	DoctorBinary string
}

func Main(w io.Writer) int {
	state, err := Load("")
	if err != nil {
		fmt.Fprintln(w, err)
		return 1
	}
	fmt.Fprintln(w, "ralphx current")
	fmt.Fprintf(w, "state_file=%s\n", state.Path)
	fmt.Fprintf(w, "version=%s\n", state.Version)
	fmt.Fprintf(w, "binary=%s\n", state.Binary)
	if strings.TrimSpace(state.DoctorBinary) != "" {
		fmt.Fprintf(w, "doctor_binary=%s\n", state.DoctorBinary)
	} else {
		fmt.Fprintln(w, "doctor_command=ralphx doctor")
	}
	return 0
}

func Load(path string) (State, error) {
	if strings.TrimSpace(path) == "" {
		path = defaultCurrentEnvPath()
	}
	f, err := os.Open(path)
	if err != nil {
		return State{}, fmt.Errorf("could not read persisted execution state: %w", err)
	}
	defer f.Close()
	state := State{Path: path}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), "\"")
		switch strings.TrimSpace(key) {
		case "RALPHX_VERSION":
			state.Version = value
		case "RALPHX_BINARY":
			state.Binary = value
		case "RALPHX_DOCTOR_BINARY":
			state.DoctorBinary = value
		}
	}
	if err := scanner.Err(); err != nil {
		return State{}, err
	}
	return state, nil
}

func defaultCurrentEnvPath() string {
	configDir := os.Getenv("RALPHX_CONFIG_DIR")
	if strings.TrimSpace(configDir) == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "~/.config/ralphx/current.env"
		}
		configDir = filepath.Join(home, ".config", "ralphx")
	}
	return filepath.Join(configDir, "current.env")
}
