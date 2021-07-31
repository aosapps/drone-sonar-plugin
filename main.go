package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/urfave/cli"
)

var version string // build number set at compile-time

func amain() {
	app := cli.NewApp()
	app.Name = "drone-sonar-scanner"
	app.Usage = "Drone plugin integrate sonar-scanner and results of SonarQube Quality Gates."
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{

		cli.StringFlag{
			Name:   "name",
			Usage:  "repository full name",
			EnvVar: "PLUGIN_NAME,DRONE_REPO",
		},
		cli.StringFlag{
			Name:   "branch",
			Usage:  "repository default branch",
			EnvVar: "PLUGIN_BRANCH,DRONE_REPO_BRANCH",
		},
		cli.StringFlag{
			Name:   "remote",
			Usage:  "git remote url",
			EnvVar: "PLUGIN_REMOTE,DRONE_REMOTE_URL",
		},
		cli.StringFlag{
			Name:   "path",
			Usage:  "git clone path",
			EnvVar: "PLUGIN_PATH,DRONE_WORKSPACE",
		},
		cli.StringFlag{
			Name:   "host",
			Usage:  "Sonar host URL",
			EnvVar: "SONAR_HOST,PLUGIN_HOST",
		},
		cli.StringFlag{
			Name:   "token",
			Usage:  "sonar token",
			EnvVar: "SONAR_TOKEN,PLUGIN_TOKEN",
		},
		cli.StringFlag{
			Name:   "key",
			Usage:  "Project Key",
			EnvVar: "PLUGIN_KEY,DRONE_REPO",
		},
		cli.StringFlag{
			Name:   "buildnum",
			Usage:  "Project version",
			EnvVar: "PLUGIN_BUILD_NUMBER,DRONE_BUILD_NUMBER",
		},
		cli.StringFlag{
			Name:   "inclusions",
			Usage:  "Project sources inclusions",
			EnvVar: "PLUGIN_INCLUSIONS",
		},
		cli.StringFlag{
			Name:   "exclusions",
			Usage:  "Project sources exclusions",
			EnvVar: "PLUGIN_EXCLUSIONS",
		},
		cli.StringFlag{
			Name:   "language",
			Usage:  "Project language",
			EnvVar: "PLUGIN_LANGUAGE",
		},
		cli.StringFlag{
			Name:   "profile",
			Usage:  "Project profile",
			EnvVar: "PLUGIN_PROFILE",
		},
		cli.StringFlag{
			Name:   "encoding",
			Usage:  "Project source encondig",
			EnvVar: "PLUGIN_ENCODING",
			Value:  "UTF-8",
		},
		cli.StringFlag{
			Name:   "quality",
			Usage:  "QualityGate status",
			EnvVar: "SONAR_QUALITYGATE,PLUGIN_QUALITYGATE",
			Value:  "OK",
		},
		cli.StringFlag{
			Name:   "settings",
			Usage:  "Project settings",
			EnvVar: "SONAR_PROJECT_SETTINGS,PLUGIN_PROJECT_SETTINGS",
			Value:  "sonar-project.properties",
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(c *cli.Context) {
	plugin := Plugin{
		Host:       c.String("host"),
		Token:      c.String("token"),
		Key:        c.String("key"),
		Name:       c.String("name"),
		Version:    c.String("buildnum"),
		Sources:    c.String("path"),
		Inclusions: c.String("inclusions"),
		Exclusions: c.String("exclusions"),
		Language:   c.String("language"),
		Profile:    c.String("profile"),
		Encoding:   c.String("encoding"),
		Remote:     c.String("remote"),
		Branch:     c.String("branch"),
		Quality:    c.String("quality"),
		Settings:   c.String("settings"),
	}
	os.Setenv("TOKEN", base64.StdEncoding.EncodeToString([]byte(c.String("token")+":")))
	if err := plugin.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
