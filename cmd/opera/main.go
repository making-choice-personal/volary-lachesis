package main

import (
	"fmt"
	"os"

	"github.com/making-choice-personal/volary-lachesis/cmd/opera/launcher"
)

func main() {
	if err := launcher.Launch(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
