package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
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
		Config Config
	}
)

type SonarStatus struct {
	ProjectStatus struct {
		Status string `json:"status"`
	}
}

func (p Plugin) getQualityGateStatus() string {
	// Check status
	// p.Config.Host /api/qualitygates/project_status?projectKey=
	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	url := fmt.Sprintf("%s/api/qualitygates/project_status?projectKey=%s", p.Config.Host, p.getProjectKey())
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	status := SonarStatus{}
	if jsonErr := json.Unmarshal(body, &status); jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return status.ProjectStatus.Status
}

func (p Plugin) getProjectKey() string {
	return strings.Replace(p.Config.Key, "/", ":", -1)
}

func (p Plugin) Exec() error {
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

	cmd := exec.Command("sonar-scanner", args...)
	// fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("==> Code Analysis Result:\n")
	err := cmd.Run()
	if err != nil {
		return err
	}

	if p.Config.EnableGateBreaker {
		qgStatus := p.getQualityGateStatus()
		fmt.Printf("==> Quality Gate status: %s\n", qgStatus)
		if status := qgStatus; status == "ERROR" {
			return fmt.Errorf("pipeline aborted because quality gate failed")
		}
	}

	return nil
}
