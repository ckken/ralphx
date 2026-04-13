package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ckken/ralphx/internal/config"
	"github.com/ckken/ralphx/internal/current"
	"github.com/ckken/ralphx/internal/doctor"
	"github.com/ckken/ralphx/internal/runner"
	"github.com/ckken/ralphx/internal/version"
)

func Main(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 0
	}

	command, rest := normalizeCommand(args)
	switch command {
	case "help", "-h", "--help":
		printUsage()
		return 0
	case "version":
		fmt.Println(version.String())
		return 0
	case "current":
		return current.Main(os.Stdout)
	case "doctor":
		return doctor.Run(os.Stdout)
	case "run":
		return run(rest)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", command)
		printUsage()
		return 1
	}
}

func run(args []string) int {
	cfg, helpShown, err := config.ParseRunArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "run argument error: %v\n", err)
		return 2
	}
	if helpShown {
		return 0
	}
	if strings.TrimSpace(cfg.TaskFile) == "" {
		fmt.Fprintln(os.Stderr, "missing required --task")
		return 2
	}

	loop := runner.New(cfg)
	if err := loop.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func normalizeCommand(args []string) (string, []string) {
	if len(args) == 0 {
		return "help", nil
	}
	first := args[0]
	switch first {
	case "run", "doctor", "version", "current", "help", "-h", "--help":
		return first, args[1:]
	default:
		if strings.HasPrefix(first, "-") {
			return "run", args
		}
		return first, args[1:]
	}
}

func printUsage() {
	fmt.Println("ralphx - lean outer-loop runner for Codex and coding agents")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ralphx run --task FILE [--checklist FILE] [--workdir DIR]")
	fmt.Println("  ralphx doctor")
	fmt.Println("  ralphx version")
	fmt.Println("  ralphx current")
	fmt.Println()
	fmt.Println("Compatibility:")
	fmt.Println("  ralphx --task FILE ...      same as: ralphx run --task FILE ...")
}
