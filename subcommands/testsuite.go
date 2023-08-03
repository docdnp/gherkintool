package subcommands

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/cucumber/godog"
)

var (
	//go:embed templates/*.robot
	templates embed.FS
)

type testSuite struct {
	godog.TestSuite
	FeatureFinder *regexp.Regexp
}

func (s *testSuite) listScenarios() {
	features, _ := s.RetrieveFeatures()
	for _, feature := range features {
		fname := s.FeatureFinder.FindStringSubmatch(feature.Uri)[1]
		for _, fchild := range feature.Feature.Children {
			fmt.Printf("Test_%v.%v\n", fname, strings.ReplaceAll(fchild.Scenario.Name, " ", "_"))
		}
	}
}

func (s *testSuite) dumpJson() {
	features, _ := s.RetrieveFeatures()
	for _, feature := range features {
		j, _ := json.Marshal(feature)
		fmt.Println(string(j))
	}
}

func (sw *testSuite) dumpRobot(tpl string, ressources []string) error {
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
			return sw.FeatureFinder.FindStringSubmatch(f)[1] + "." + strings.ReplaceAll(cs, " ", "_")
		},
		"purename":   func(s string) string { return sw.FeatureFinder.FindStringSubmatch(s)[1] },
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
