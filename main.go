package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/lnutimura/eol/cmd"
)

// version is set at build time by GoReleaser.
var version = "dev"

func main() {
	if err := cmd.Execute(version); err != nil {
		var ee *cmd.ExitError
		if errors.As(err, &ee) {
			if ee.Err != nil {
				fmt.Fprintln(os.Stderr, ee.Err)
			}
			os.Exit(ee.Code)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
