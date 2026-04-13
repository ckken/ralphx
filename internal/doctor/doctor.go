package doctor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Check struct {
	Name     string
	Required bool
}

func Main(args []string) int {
	return Run(os.Stdout)
}

func Run(w io.Writer) int {
	fmt.Fprintln(w, "ralphx doctor")
	fmt.Fprintf(w, "BIN_DIR=%s\n\n", defaultBinDir())

	checks := []Check{
		{Name: "bash", Required: true},
		{Name: "python3", Required: true},
		{Name: "git", Required: false},
		{Name: "gh", Required: false},
		{Name: "codex", Required: true},
		{Name: "jq", Required: false},
	}

	exitCode := 0
	for _, check := range checks {
		if err := printCheck(w, check); err != nil && check.Required {
			exitCode = 1
		}
	}

	fmt.Fprintln(w)
	pathEnv := os.Getenv("PATH")
	if strings.Contains(":"+pathEnv+":", ":"+defaultBinDir()+":") {
		fmt.Fprintf(w, "[ok] PATH contains %s\n", defaultBinDir())
	} else {
		fmt.Fprintf(w, "[missing] PATH does not contain %s\n", defaultBinDir())
		fmt.Fprintf(w, "Add it with: export PATH=\"%s:$PATH\"\n", defaultBinDir())
	}

	return exitCode
}

func printCheck(w io.Writer, check Check) error {
	path, err := exec.LookPath(check.Name)
	if err != nil {
		if check.Required {
			fmt.Fprintf(w, "[missing] %s (required)\n", check.Name)
		} else {
			fmt.Fprintf(w, "[missing] %s (optional)\n", check.Name)
		}
		return err
	}
	fmt.Fprintf(w, "[ok] %s -> %s\n", check.Name, path)
	return nil
}

func defaultBinDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.local/bin"
	}
	return home + "/.local/bin"
}
