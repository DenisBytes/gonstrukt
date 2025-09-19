package main

import (
	"fmt"
	"os"

	"github.com/DenisBytes/gonstrukt/cmd"
)

func main() {
	if err := cmd.RootCmd().Execute(); err != nil {
		// Use custom error formatting
		fmt.Fprint(os.Stderr, cmd.FormatCliError(err))
		os.Exit(1)
	}
}