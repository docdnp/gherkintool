package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/docdnp/gherkintool/subcommands"
)

func main() {
	var cmd string
	os.Args = os.Args[1:]
	if len(os.Args) >= 1 {
		cmd = os.Args[0]
	}

	gherkintool := subcommands.NewFlagSet("gherkintool")

	switch cmd {
	case "feature":
		subcommands.Feature(os.Args)
	case "robot":
		subcommands.Robot(os.Args)
	case "":
		gherkintool.Error(fmt.Errorf("choose a sub command"))
	default:
		gherkintool.Error(fmt.Errorf("unknown sub command: %v", cmd))
	}

	os.Exit(0)
}
