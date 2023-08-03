package subcommands

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/pflag"
)

var (
	//go:embed templates/usage.txt
	usageIntro string
)

var allFlagSets []*FlagSet = make([]*FlagSet, 0)

type FlagSet struct {
	*pflag.FlagSet
	// Error     func(err error)
	// UsageExit func()
	Name string
}

func (f *FlagSet) Error(err error) {
	reindent := regexp.MustCompile(`(?is)(\s*)(-)(.*?\n)`)
	fmt.Printf("error: %v\n", err)
	fmt.Print(usageIntro)
	for _, f := range allFlagSets {
		f.dumpUsage(reindent)
	}
	os.Exit(1)
}

func NewFlagSet(name string) *FlagSet {
	fs := &FlagSet{FlagSet: pflag.NewFlagSet(name, pflag.ContinueOnError), Name: name}
	allFlagSets = append(allFlagSets, fs)
	fs.FlagSet.Usage = Usage()
	// fs.UsageExit = func() { fs.Usage(); os.Exit(0) }
	// fs.Error = func(err error) { fmt.Printf("%v\n", err); fs.Usage(); os.Exit(1) }
	return fs
}

var Usage = func() func() {
	reindent := regexp.MustCompile(`(?is)(\s*)(-)(.*?\n)`)
	return func() {
		fmt.Print(usageIntro)
		for _, f := range allFlagSets {
			f.dumpUsage(reindent)
		}
		os.Exit(0)
	}
}

func (fs *FlagSet) dumpUsage(reindent *regexp.Regexp) {
	if !fs.HasFlags() {
		return
	}
	h := fmt.Sprintf("-%s: \n", fs.Name)
	fmt.Print(
		reindent.ReplaceAllString(h, "    $1$3"),
		reindent.ReplaceAllString(fs.FlagUsages(), "     $1$2$3"),
		"\n",
	)
}
