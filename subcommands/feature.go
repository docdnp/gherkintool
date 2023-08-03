package subcommands

import (
	"fmt"
	"os"
)

func createFeatures(filenames []string) error {
	if len(filenames) == 0 {
		return fmt.Errorf("specify a code file containing comments to extract")
	}
	intro, out := func(s string) {}, func(s string) {}
	if len(filenames) > 1 {
		intro = func(s string) { fmt.Print("=== ", s, "\n") }
		out = func(s string) { fmt.Print("=== ", s, "\n\n") }
	}
	dparser := newDslParser()
	for _, filename := range filenames {
		file, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("unknown file '%s': %v", filename, err)
		}
		features := dparser.extractFeatures(filename, string(file))

		if len(features) == 0 {
			continue
		}
		intro(filename)
		for _, i := range features {
			fmt.Printf("%s", i)
		}
		out(filename)
	}
	return nil
}

func Feature(args []string) {
	flags := NewFlagSet("feature")
	if len(args) == 1 {
		flags.Error(fmt.Errorf("need at least one file as argument"))
	}
	if err := flags.Parse(args); err != nil {
		flags.Error(err)
	}
	err := createFeatures(os.Args[1:])
	if err != nil {
		flags.Error(err)
	} else {
		os.Exit(0)
	}
	createFeatures(args)
	os.Exit(0)
}
