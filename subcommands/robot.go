package subcommands

import (
	"fmt"
	"os"
	"regexp"

	"github.com/cucumber/godog"
	_ "github.com/spf13/pflag"
)

func Robot(args []string) {
	flags := NewFlagSet("robot")
	cfg := struct {
		useTplScenario *bool
		listScenarios  *bool
		dumpJson       *bool
		features       *[]string
		ressources     *[]string
		tags           *string
	}{
		useTplScenario: flags.BoolP("tpl.feature", "F", false,
			"create one robot test per scenario."),
		listScenarios: flags.BoolP("list", "l", false,
			"list all available scenarios as they are named during a godog run"),
		dumpJson: flags.BoolP("json", "j", false,
			"dump all features as JSON"),
		features: flags.StringSliceP("features", "f", []string{"./features"},
			"provide a list of directories or feature files"),
		ressources: flags.StringSliceP("resources", "r", nil,
			"provide ressources to be included in robot file"),
		tags: flags.StringP("tags", "t", "~@wip",
			"specify tags to include or exclude features or steps"),
	}

	if len(args) == 1 {
		flags.Error(fmt.Errorf("no arguments given"))
	}
	errs := []error{}

	flags.Parse(args)

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
		os.Exit(1)
	}

	suite := testSuite{
		TestSuite: godog.TestSuite{
			ScenarioInitializer: func(sc *godog.ScenarioContext) {},
			Options: &godog.Options{
				Paths: *cfg.features,
				Tags:  *cfg.tags,
			},
		},
		FeatureFinder: regexp.MustCompile(`.*/(.*)\.feature$`),
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
		flags.Error(err)
	}
	os.Exit(0)

}
