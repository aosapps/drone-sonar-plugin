package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type (
	Config struct {
		Key   string
		Name  string
		Host  string
		Token string

		Version        string
		Branch         string
		Sources        string
		Timeout        string
		Inclusions     string
		Exclusions     string
		Level          string
		ShowProfiling  string
		BranchAnalysis bool
		UsingProperties bool
	}
	Plugin struct {
		Config Config
	}
)

func (p Plugin) Exec() error {
	args := []string{
		"-Dsonar.host.url=" + p.Config.Host,
		"-Dsonar.login=" + p.Config.Token,
	}

	if !p.Config.UsingProperties {
		argsParameter := []string{
			"-Dsonar.projectKey=" + strings.Replace(p.Config.Key, "/", ":", -1),
			"-Dsonar.projectName=" + p.Config.Name,
			"-Dsonar.projectVersion=" + p.Config.Version,
			"-Dsonar.sources=" + p.Config.Sources,
			"-Dsonar.ws.timeout=" + p.Config.Timeout,
			"-Dsonar.inclusions=" + p.Config.Inclusions,
			"-Dsonar.exclusions=" + p.Config.Exclusions,
			"-Dsonar.log.level=" + p.Config.Level,
			"-Dsonar.showProfiling=" + p.Config.ShowProfiling,
			"-Dsonar.scm.provider=git",
		}
		args = append(args, argsParameter...)
	}


	if p.Config.BranchAnalysis {
		args = append(args, "-Dsonar.branch.name=" + p.Config.Branch)
	}

	cmd := exec.Command("sonar-scanner", args...)
	// fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("==> Code Analysis Result:\n")
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
