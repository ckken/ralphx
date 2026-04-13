package main

import (
	"os"

	"github.com/ckken/ralphx/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}
