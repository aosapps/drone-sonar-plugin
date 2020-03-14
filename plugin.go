package main

import (
	"bytes"
	"fmt"
	sonargo "github.com/magicsong/sonargo/sonar"
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

		Version           string
		Branch            string
		Sources           string
		Timeout           string
		Inclusions        string
		Exclusions        string
		Level             string
		ShowProfiling     string
		BranchAnalysis    bool
		UsingProperties   bool
		EnableGateBreaker bool
	}
	Plugin struct {
		Config      Config
		SonarClient *sonargo.Client
	}
)

func (p Plugin) getProjectKey() string {
	return strings.Replace(p.Config.Key, "/", ":", -1)
}

// Returns array of arguments that will be used during the command call
func (p Plugin) getCommandArgs() []string {
	args := []string{
		"-Dsonar.host.url=" + p.Config.Host,
		"-Dsonar.login=" + p.Config.Token,
	}

	if !p.Config.UsingProperties {
		argsParameter := []string{
			"-Dsonar.projectKey=" + p.getProjectKey(),
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

	return args
}

func (p Plugin) Exec() error {
	args := p.getCommandArgs()
	cmd := exec.Command("sonar-scanner", args...)
	// fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	fmt.Printf("==> Code Analysis Result:\n")
	err := cmd.Run()
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(outb.Bytes())
	_, _ = os.Stderr.Write(errb.Bytes())

	if p.Config.EnableGateBreaker {
		// Extract task id from command log
		taskId, extractError := p.extractReportIdFromAnalysisLog(outb.String())
		if extractError != nil {
			return extractError
		}
		// Check if quality gate succeeded
		if err = p.validateQualityGate(taskId); err != nil {
			return err
		}
	}

	return nil
}

func NewPlugin(config Config) (*Plugin, error) {
	client, err := sonargo.NewClientByToken(config.Host+"/api", config.Token)
	if err != nil {
		return nil, err
	}

	return &Plugin{
		Config:      config,
		SonarClient: client,
	}, nil
}
