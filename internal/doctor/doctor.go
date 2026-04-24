package doctor

import (
	"encoding/json"
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

type CheckResult struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Found    bool   `json:"found"`
	Path     string `json:"path,omitempty"`
}

type Result struct {
	Tool               string        `json:"tool"`
	BinDir             string        `json:"bin_dir"`
	Checks             []CheckResult `json:"checks"`
	PathContainsBinDir bool          `json:"path_contains_bin_dir"`
	Ok                 bool          `json:"ok"`
	ExitCode           int           `json:"exit_code"`
}

func Main(args []string) int {
	return run(args, os.Stdout, os.Stderr)
}

func run(args []string, stdout, stderr io.Writer) int {
	jsonMode := false
	help := false
	extra := make([]string, 0, len(args))
	for _, arg := range args {
		switch arg {
		case "--json":
			jsonMode = true
		case "--help", "-h":
			help = true
		case "help":
			help = true
		default:
			extra = append(extra, arg)
		}
	}

	if len(extra) > 0 {
		if jsonMode {
			return writeJSON(stdout, Result{
				Tool:     "doctor",
				BinDir:   defaultBinDir(),
				Ok:       false,
				ExitCode: 2,
			}, "unexpected arguments: "+strings.Join(extra, " "))
		}
		fmt.Fprintf(stderr, "doctor argument error: unexpected arguments: %s\n", strings.Join(extra, " "))
		return 2
	}

	if help {
		printUsage(stdout)
		return 0
	}

	result := collectResult()
	if jsonMode {
		if err := json.NewEncoder(stdout).Encode(result); err != nil {
			fmt.Fprintf(stderr, "doctor json output error: %v\n", err)
			return 1
		}
		return result.ExitCode
	}

	printText(stdout, result)
	return result.ExitCode
}

func collectResult() Result {
	checks := []Check{
		{Name: "bash", Required: true},
		{Name: "python3", Required: true},
		{Name: "git", Required: false},
		{Name: "gh", Required: false},
		{Name: "codex", Required: true},
		{Name: "jq", Required: false},
	}

	result := Result{
		Tool:   "doctor",
		BinDir: defaultBinDir(),
	}
	for _, check := range checks {
		entry := CheckResult{Name: check.Name, Required: check.Required}
		path, err := exec.LookPath(check.Name)
		if err == nil {
			entry.Found = true
			entry.Path = path
		}
		if err != nil && check.Required {
			result.ExitCode = 1
		}
		result.Checks = append(result.Checks, entry)
	}

	pathEnv := os.Getenv("PATH")
	result.PathContainsBinDir = strings.Contains(":"+pathEnv+":", ":"+result.BinDir+":")
	result.Ok = result.ExitCode == 0
	return result
}

func printText(w io.Writer, result Result) {
	fmt.Fprintln(w, "ralphx doctor")
	fmt.Fprintf(w, "BIN_DIR=%s\n\n", result.BinDir)
	for _, check := range result.Checks {
		if check.Found {
			fmt.Fprintf(w, "[ok] %s -> %s\n", check.Name, check.Path)
			continue
		}
		if check.Required {
			fmt.Fprintf(w, "[missing] %s (required)\n", check.Name)
			continue
		}
		fmt.Fprintf(w, "[missing] %s (optional)\n", check.Name)
	}

	fmt.Fprintln(w)
	if result.PathContainsBinDir {
		fmt.Fprintf(w, "[ok] PATH contains %s\n", result.BinDir)
	} else {
		fmt.Fprintf(w, "[missing] PATH does not contain %s\n", result.BinDir)
		fmt.Fprintf(w, "Add it with: export PATH=\"%s:$PATH\"\n", result.BinDir)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: ralphx doctor [--json]")
	fmt.Fprintln(w, "  --json   emit machine-readable JSON output")
}

func writeJSON(w io.Writer, result Result, errMsg string) int {
	payload := map[string]any{
		"tool":      result.Tool,
		"bin_dir":   result.BinDir,
		"ok":        result.Ok,
		"exit_code": result.ExitCode,
		"error":     errMsg,
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(payload); err != nil {
		return 1
	}
	if result.ExitCode != 0 {
		return result.ExitCode
	}
	return 0
}

func defaultBinDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.local/bin"
	}
	return home + "/.local/bin"
}
