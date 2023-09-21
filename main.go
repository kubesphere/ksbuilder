package main

import (
	"fmt"
	"os"

	"github.com/kubesphere/ksbuilder/cmd"
)

// This value will be injected into the corresponding git tag value at build time using `-ldflags`.
var version = "0.0.0"

func main() {
	if err := cmd.Execute(version); err != nil {
		fmt.Fprintf(os.Stderr, "\n%v\n", err)
		os.Exit(1)
	}
}
