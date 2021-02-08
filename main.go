package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

var build = "1" // build number set at compile time

func main() {
	app := cli.NewApp()
	app.Name = "Drone-Sonar-Plugin"
	app.Usage = "Drone plugin to integrate with SonarQube."
	app.Action = run
	app.Version = fmt.Sprintf("1.0.%s", build)
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "key",
			Usage:   "project key",
			EnvVars: []string{"DRONE_REPO"},
		},
		&cli.StringFlag{
			Name:    "name",
			Usage:   "project name",
			EnvVars: []string{"DRONE_REPO"},
		},
		&cli.StringFlag{
			Name:    "host",
			Usage:   "SonarQube host",
			EnvVars: []string{"PLUGIN_SONAR_HOST"},
		},
		&cli.StringFlag{
			Name:    "token",
			Usage:   "SonarQube token",
			EnvVars: []string{"PLUGIN_SONAR_TOKEN"},
		},

		// advanced parameters
		&cli.StringFlag{
			Name:    "ver",
			Usage:   "Project version",
			EnvVars: []string{"DRONE_BUILD_NUMBER"},
		},
		&cli.StringFlag{
			Name:    "branch",
			Usage:   "Project branch",
			EnvVars: []string{"DRONE_BRANCH"},
		},
		&cli.StringFlag{
			Name:    "timeout",
			Usage:   "Web request timeout",
			Value:   "60",
			EnvVars: []string{"PLUGIN_TIMEOUT"},
		},
		&cli.StringFlag{
			Name:    "sources",
			Usage:   "analysis sources",
			Value:   ".",
			EnvVars: []string{"PLUGIN_SOURCES"},
		},
		&cli.StringFlag{
			Name:    "inclusions",
			Usage:   "code inclusions",
			EnvVars: []string{"PLUGIN_INCLUSIONS"},
		},
		&cli.StringFlag{
			Name:    "exclusions",
			Usage:   "code exclusions",
			EnvVars: []string{"PLUGIN_EXCLUSIONS"},
		},
		&cli.StringFlag{
			Name:    "level",
			Usage:   "log level",
			Value:   "INFO",
			EnvVars: []string{"PLUGIN_LEVEL"},
		},
		&cli.StringFlag{
			Name:    "showProfiling",
			Usage:   "showProfiling during analysis",
			Value:   "false",
			EnvVars: []string{"PLUGIN_SHOWPROFILING"},
		},
		&cli.BoolFlag{
			Name:    "branchAnalysis",
			Usage:   "execute branchAnalysis",
			EnvVars: []string{"PLUGIN_BRANCHANALYSIS"},
		},
		&cli.BoolFlag{
			Name:    "usingProperties",
			Usage:   "using sonar-project.properties",
			EnvVars: []string{"PLUGIN_USINGPROPERTIES"},
		},
		&cli.BoolFlag{
			Name:    "trustServerCert",
			Usage:   "trust sonar server certificate",
			EnvVars: []string{"PLUGIN_TRUSTSERVERCERT"},
		},
	}

	app.Run(os.Args)
}

func run(c *cli.Context) error {
	plugin := Plugin{
		Config: Config{
			Key:   c.String("key"),
			Name:  c.String("name"),
			Host:  c.String("host"),
			Token: c.String("token"),

			Version:         c.String("ver"),
			Branch:          c.String("branch"),
			Timeout:         c.String("timeout"),
			Sources:         c.String("sources"),
			Inclusions:      c.String("inclusions"),
			Exclusions:      c.String("exclusions"),
			Level:           c.String("level"),
			ShowProfiling:   c.String("showProfiling"),
			BranchAnalysis:  c.Bool("branchAnalysis"),
			UsingProperties: c.Bool("usingProperties"),
			TrustServerCert: c.Bool("trustServerCert"),
		},
	}

	if plugin.Config.TrustServerCert {
		if err := plugin.TrustServerCert(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	if err := plugin.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}
