package main

import (
	"os"

	"github.com/ckken/ralphx/internal/doctor"
)

func main() {
	os.Exit(doctor.Main(os.Args[1:]))
}
