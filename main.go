package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/cucumber/godog"
	"github.com/docdnp/gherkintool/subcommands"
	"github.com/spf13/pflag"
)

var (
	//go:embed templates/*.robot
	templates embed.FS

	//go:embed templates/usage.txt
	usageIntro string
)

type godogWrapper struct {
	godog.TestSuite
	re_feature *regexp.Regexp
}

func (s *godogWrapper) listScenarios() {
	features, _ := s.RetrieveFeatures()
	for _, feature := range features {
		fname := s.re_feature.FindStringSubmatch(feature.Uri)[1]
		for _, fchild := range feature.Feature.Children {
			fmt.Printf("Test_%v.%v\n", fname, strings.ReplaceAll(fchild.Scenario.Name, " ", "_"))
		}
	}
}

func (s *godogWrapper) dumpJson() {
	features, _ := s.RetrieveFeatures()
	for _, feature := range features {
		j, _ := json.Marshal(feature)
		fmt.Println(string(j))
	}
}

func (sw *godogWrapper) dumpRobot(tpl string, ressources []string) error {
	m1 := regexp.MustCompile(`\s+`)

	features, err := sw.RetrieveFeatures()
	if err != nil {
		return err
	}
	jfeatures := make([]interface{}, 0, len(features))
	for _, feature := range features {
		j, _ := json.Marshal(feature)
		var jf map[string]interface{}
		json.Unmarshal(j, &jf)
		jfeatures = append(jfeatures, jf)
	}

	fm := template.FuncMap{
		"Title": func(f string, cs string) string {
			return sw.re_feature.FindStringSubmatch(f)[1] + "." + strings.ReplaceAll(cs, " ", "_")
		},
		"purename":   func(s string) string { return sw.re_feature.FindStringSubmatch(s)[1] },
		"puretag":    func(s string) string { return strings.Replace(s, "@", "", -1) },
		"features":   func() []interface{} { return jfeatures },
		"cleantitle": func(s string) string { return strings.ReplaceAll(s, "'", "`") },
		"rmnewln": func(s string) string {
			return m1.ReplaceAllString(
				strings.ReplaceAll(s, "\n", ""),
				" ")
		},
	}

	t, err := template.New("main.robot").Funcs(fm).ParseFS(templates, "*/*.robot")
	if err != nil {
		panic(err)
	}

	data := struct {
		Features   []interface{}
		Ressources []string
		Template   string
	}{jfeatures, ressources, tpl}

	var tplx bytes.Buffer
	if err = t.Execute(&tplx, data); err != nil {
		panic(err)
	}
	fmt.Println(tplx.String())
	return nil
}

var usage = func(fs ...*flagSet) func() {
	reindent := regexp.MustCompile(`(?is)(\s*)(-)(.*?\n)`)
	return func() {
		fmt.Print(usageIntro)
		for _, f := range fs {
			f.dumpUsage(reindent)
		}
	}
}

func (fs *flagSet) dumpUsage(reindent *regexp.Regexp) {
	h := fmt.Sprintf("-%s: \n", fs.Name)
	fmt.Print(
		reindent.ReplaceAllString(h, "    $1$3"),
		reindent.ReplaceAllString(fs.FlagUsages(), "     $1$2$3"),
		"\n",
	)
}

type flagSet struct {
	*pflag.FlagSet
	Error func(err error)
	Name  string
}

func newFlagSet(name string) *flagSet {
	return &flagSet{FlagSet: pflag.NewFlagSet(name, pflag.ContinueOnError), Name: name}
}

func main() {
	robot := newFlagSet("robot")
	feature := newFlagSet("feature")
	robot.Usage = usage(robot)
	feature.Usage = func() { robot.Usage(); os.Exit(0) }
	feature.Error = func(err error) { fmt.Println("Error:", err); robot.Usage(); os.Exit(1) }

	cfg := struct {
		useTplScenario *bool
		listScenarios  *bool
		dumpJson       *bool
		features       *[]string
		ressources     *[]string
		tags           *string
	}{
		useTplScenario: robot.BoolP("tpl.feature", "F", false,
			"create one robot test per scenario."),
		listScenarios: robot.BoolP("list", "l", false,
			"list all available scenarios as they are named during a godog run"),
		dumpJson: robot.BoolP("json", "j", false,
			"dump all features as JSON"),
		features: robot.StringSliceP("features", "f", []string{"./features"},
			"provide a list of directories or feature files"),
		ressources: robot.StringSliceP("resources", "r", nil,
			"provide ressources to be included in robot file"),
		tags: robot.StringP("tags", "t", "~@wip",
			"specify tags to include or exclude features or steps"),
	}

	var cmd string
	errs := []error{}
	os.Args = os.Args[1:]
	if len(os.Args) >= 1 {
		cmd = os.Args[0]
	}

	switch cmd {
	case "feature":
		if err := feature.Parse(os.Args); err != nil {
			feature.Error(err)
		}
		err := subcommands.Feature(os.Args[1:])
		if err != nil {
			errs = append(errs, err)
		} else {
			os.Exit(0)
		}
	case "robot":
		robot.Parse(os.Args[1:])
		if len(*cfg.features) == 0 {
			errs = append(errs, fmt.Errorf("didn't specify feature directories of files"))
		}
	case "":
		errs = append(errs, fmt.Errorf("specify a subcommand"))
	default:
		errs = append(errs, fmt.Errorf("unknown subcommand: '%s'", cmd))
	}

	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Println("Error:", e)
		}
		robot.Usage()
		os.Exit(1)
	}

	for _, f := range *cfg.features {
		if _, e := os.Stat(f); os.IsNotExist(e) {
			if e == nil {
				continue
			}
			errs = append(errs, fmt.Errorf("unknown feature dir or file: %w", e))
		}
	}

	suite := godogWrapper{godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {},
		Options: &godog.Options{
			Paths: *cfg.features,
			Tags:  *cfg.tags,
		},
	},
		regexp.MustCompile(`.*/(.*)\.feature$`),
	}

	if *cfg.listScenarios {
		suite.listScenarios()
		os.Exit(0)
	}
	if *cfg.dumpJson {
		suite.dumpJson()
		os.Exit(0)
	}

	tpl := "features"
	if *cfg.useTplScenario {
		tpl = "scenarios"
	}

	if err := suite.dumpRobot(tpl, *cfg.ressources); err != nil {
		feature.Error(err)
	}
	os.Exit(0)

}
