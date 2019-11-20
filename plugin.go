package main

import (
	"fmt"
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
		showProfiling  string
		branchAnalysis bool
	}
	Plugin struct {
		Config Config
	}
)

func (p Plugin) Exec() error {
	args := []string{
		"-Dsonar.projectKey=" + strings.Replace(p.Config.Key, "/", ":", -1),
		"-Dsonar.projectName=" + p.Config.Name,
		"-Dsonar.host.url=" + p.Config.Host,
		"-Dsonar.login=" + p.Config.Token,

		"-Dsonar.projectVersion=" + p.Config.Version,
		"-Dsonar.sources=" + p.Config.Sources,
		"-Dsonar.ws.timeout=" + p.Config.Timeout,
		"-Dsonar.inclusions=" + p.Config.Inclusions,
		"-Dsonar.exclusions=" + p.Config.Exclusions,
		"-Dsonar.log.level=" + p.Config.Level,
		"-Dsonar.showProfiling=" + p.Config.showProfiling,
		"-Dsonar.scm.provider=git",
	}

	if p.Config.branchAnalysis {
		args = append(args, "-Dsonar.branch.name="+p.Config.Branch)
	}

	cmd := exec.Command("sonar-scanner", args...)
	// fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Printf("==> Code Analysis Result: %s\n", string(output))
	}
	if err != nil {
		return err
	}

	return nil
}
