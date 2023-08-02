package subcommands

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
)

type features = []string
type dslRexp = struct {
	*regexp.Regexp
	name    string
	replace string
}
type dslParser struct {
	blockFinder    dslRexp
	featureFinder  dslRexp
	scenarioTagger dslRexp
	replacer       []dslRexp
}

func newDslParser() *dslParser {
	return &dslParser{
		blockFinder:    dslRexp{Regexp: regexp.MustCompile(`(?is)\[gherkin\].*?\n(.*?)\[end\]\n(?:func\s+(.*?)\()?`)},
		featureFinder:  dslRexp{Regexp: regexp.MustCompile(`(?s)(?:@.*?\n)*\nFeature`)},
		scenarioTagger: dslRexp{Regexp: regexp.MustCompile(`^(\s*)`)},
		replacer: []dslRexp{
			{name: "CommentToken", replace: "$1",
				Regexp: regexp.MustCompile(`(?s)//(\s+)`),
			},
			{name: "Tabs", replace: "    ",
				Regexp: regexp.MustCompile(`\t`),
			},
			{name: "BlankAtLineStart", replace: "$1",
				Regexp: regexp.MustCompile(`(?s)(^|\n) `),
			},
			{name: "EmptyLines", replace: "\n",
				Regexp: regexp.MustCompile(`(?s)\n+`),
			},
		},
	}
}

func (d *dslParser) extractFeatures(filename string, content string) features {
	features := make(features, 0)
	match := d.blockFinder.FindAllStringSubmatch(content, -1)
	for _, m := range match {
		rm := make([]string, len(m)-1)
		copy(rm, m[1:])
		sort.SliceStable(rm, func(i, j int) bool { return i > j })
		if len(rm) < 2 {
			log.Printf("error: %s: too short: %v", os.Args[1], rm)
			break
		}

		var cleaned = rm[1]
		for _, replacer := range d.replacer {
			cleaned = replacer.ReplaceAllString(cleaned, replacer.replace)
		}

		if d.featureFinder.MatchString(cleaned) {
			features = append(features, cleaned)
			continue
		}
		cleaned = d.scenarioTagger.ReplaceAllString(cleaned, fmt.Sprintf("$1@file:%s$1@test:%s$1", filename, rm[0]))
		features = append(features, cleaned)
	}

	return features
}

func Feature(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("unknown file '%s': %v", filename, err)
	}
	features := newDslParser().extractFeatures(filename, string(file))

	for _, i := range features {
		fmt.Printf("%s", i)
	}
	return nil
}
