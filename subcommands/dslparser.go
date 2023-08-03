package subcommands

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

type features = []string
type dslRexp = struct {
	*regexp.Regexp
	name    string
	replace string
}
type dslParser struct {
	featureFinder       dslRexp
	scenarioFinder      dslRexp
	scenarioTagger      dslRexp
	scenarioTitleFinder dslRexp
	replacer            []dslRexp
}

func newDslParser() *dslParser {
	return &dslParser{
		featureFinder:       dslRexp{Regexp: regexp.MustCompile(`(?s)(?:@.*?\n)*\nFeature`)},
		scenarioFinder:      dslRexp{Regexp: regexp.MustCompile(`(?is)\[gherkin\].*?\n(.*?)\[end\]\n(?:func\s+(.*?)\()?`)},
		scenarioTagger:      dslRexp{Regexp: regexp.MustCompile(`^(\s*)`)},
		scenarioTitleFinder: dslRexp{Regexp: regexp.MustCompile(`(?is)Scenario:\s*\n`)},
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

func (*dslParser) makeScenarioTitle(testFunc string) string {
	scnTitle := strings.Replace(testFunc, "Test_", "", -1)
	scnTitle = "Scenario: " + strings.Replace(scnTitle, "_", " ", -1)
	return scnTitle
}

func (d *dslParser) extractFeatures(filename string, content string) features {
	features := make(features, 0)
	match := d.scenarioFinder.FindAllStringSubmatch(content, -1)
	for _, m := range match {
		rm := make([]string, len(m)-1)
		copy(rm, m[1:])
		sort.SliceStable(rm, func(i, j int) bool { return i > j })
		if len(rm) < 2 {
			log.Printf("error: %s: too short: %v", os.Args[1], rm)
			break
		}

		testFunc := rm[0]
		scenario := rm[1]

		scenario = d.scenarioTitleFinder.ReplaceAllString(scenario, d.makeScenarioTitle(testFunc))

		for _, replacer := range d.replacer {
			scenario = replacer.ReplaceAllString(scenario, replacer.replace)
		}

		if d.featureFinder.MatchString(scenario) {
			features = append(features, scenario)
			continue
		}
		scenario = d.scenarioTagger.ReplaceAllString(scenario, fmt.Sprintf("$1@file:%s$1@test:%s$1", filename, testFunc))
		features = append(features, scenario)
	}

	return features
}
