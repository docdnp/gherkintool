package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/cucumber/godog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
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

var tplPerScenario = `{{- if .Ressources}}*** Settings ***
{{- range $r := .Ressources  }}
Resource  {{$r}}
{{- end}}

{{else}}
{{- end}}*** Test Cases ***
{{- range $feature := .Features }}
{{- range $scen_idx, $child := $feature.feature.children }}
{{ purename $feature.uri }}= {{ cleantitle $child.scenario.name }}
    [Documentation]    {{ rmnewln $feature.feature.description  }}
    [Tags]    {{range $tag := $child.scenario.tags }}{{ puretag $tag.name }}    {{end}}{{ $feature.uri }}
    Perform Scenario {{ purename $feature.uri }}: {{ $feature.feature.name }}
{{range $s := $child.scenario.steps }}    {{$s.keyword}}{{$s.text}}
{{end}}
{{- end}}
{{- end}}
`

var tplPerFeature = `{{- if .Ressources}}*** Settings ***
{{- range $r := .Ressources  }}
Resource  {{$r}}
{{- end}}

{{else}}
{{- end}}*** Test Cases ***
{{- range $feature := .Features }}
{{ cleantitle $feature.feature.name }}
    [Documentation]    {{ rmnewln $feature.feature.description  }}
    [Tags]    {{range $tag := $feature.feature.tags }}{{ puretag $tag.name }}    {{end}}{{ $feature.uri }}
{{- range $scen_idx, $child := $feature.feature.children }}
    Perform Scenario {{purename $feature.uri }}: {{ $child.scenario.name }}
{{- range $step_idx, $step := $child.scenario.steps }}
    {{$step.keyword}}{{$step.text}}
{{- end}}
{{- end}}
{{end}}
`

func (sw *godogWrapper) dumpRobot(tpl string, ressources []string) {
	m1 := regexp.MustCompile(`\s+`)

	features, _ := sw.RetrieveFeatures()
	jfeatures := make([]interface{}, 0, len(features))
	for _, feature := range features {
		j, _ := json.Marshal(feature)
		var jf map[string]interface{}
		json.Unmarshal(j, &jf)
		jfeatures = append(jfeatures, jf)
	}

	t := template.Must(template.New("").
		Funcs(template.FuncMap{
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
		},
		).
		Parse(tpl))

	var tplx bytes.Buffer
	t.Execute(&tplx, struct {
		Features   []interface{}
		Ressources []string
	}{jfeatures, ressources})
	fmt.Println(tplx.String())
}

func main() {
	cfg := struct {
		useTplScenario *bool
		listScenarios  *bool
		dumpJson       *bool
		features       *[]string
		ressources     *[]string
		tags           *string
	}{
		useTplScenario: pflag.BoolP("tpl.feature", "F", false,
			"create one robot test per scenario."),
		listScenarios: pflag.BoolP("list", "l", false,
			"list all available scenarios as they are named during a godog run"),
		dumpJson: pflag.BoolP("json", "j", false,
			"dump all features as JSON"),
		features: pflag.StringSliceP("features", "f", []string{"./features"},
			"provide a list of directories or feature files"),
		ressources: pflag.StringSliceP("resources", "r", nil,
			"provide ressources to be included in robot file"),
		tags: pflag.StringP("tags", "t", "~@wip",
			"specify tags to include or exclude features or steps"),
	}
	pflag.Parse()

	errs := []error{}
	if len(*cfg.features) == 0 {
		errs = append(errs, fmt.Errorf("didn't specify feature directories of files"))
	}
	for _, f := range *cfg.features {
		if _, e := os.Stat(f); os.IsNotExist(e) {
			if e == nil {
				continue
			}
			errs = append(errs, fmt.Errorf("unknown feature dir or file: %w", e))
		}
	}
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Println("Error:", e)
		}
		pflag.Usage()
		os.Exit(1)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

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

	tpl := tplPerFeature
	if *cfg.useTplScenario {
		tpl = tplPerScenario
	}

	suite.dumpRobot(tpl, *cfg.ressources)
	os.Exit(0)

}
