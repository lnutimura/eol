package main

import (
	"fmt"
	"os"

	"github.com/lutimura/eol/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred while executing eol: %s\n", err)
		os.Exit(1)
	}
}
