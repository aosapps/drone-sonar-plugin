package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
)

var netClient *http.Client

type (
	Config struct {
		Key   string
		Name  string
		Host  string
		Token string

		Version         string
		Branch          string
		Sources         string
		Timeout         string
		Inclusions      string
		Exclusions      string
		Level           string
		ShowProfiling   string
		BranchAnalysis  bool
		UsingProperties bool
		Binaries        string
		Quality         string
		Settings        string
		Language        string
		Profile         string
		Encoding        string
		Remote          string
	}
	Plugin struct {
		Config Config
	}
	// SonarReport it is the representation of .scannerwork/report-task.txt
	SonarReport struct {
		ProjectKey   string `toml:"projectKey"`
		ServerURL    string `toml:"serverUrl"`
		DashboardURL string `toml:"dashboardUrl"`
		CeTaskID     string `toml:"ceTaskId"`
		CeTaskURL    string `toml:"ceTaskUrl"`
	}
	// TaskResponse Give Compute Engine task details such as type, status, duration and associated component.
	TaskResponse struct {
		Task struct {
			ID            string `json:"id"`
			Type          string `json:"type"`
			ComponentID   string `json:"componentId"`
			ComponentKey  string `json:"componentKey"`
			ComponentName string `json:"componentName"`
			AnalysisID    string `json:"analysisId"`
			Status        string `json:"status"`
		} `json:"task"`
	}
	// ProjectStatusResponse Get the quality gate status of a project or a Compute Engine task
	ProjectStatusResponse struct {
		ProjectStatus struct {
			Status string `json:"status"`
		} `json:"projectStatus"`
	}
)

func init() {
	netClient = &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}
}

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
			"-Dsonar.java.binaries=" + p.Config.Binaries,
		}
		args = append(args, argsParameter...)
	}

	if p.Config.BranchAnalysis {
		args = append(args, "-Dsonar.branch.name="+p.Config.Branch)
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
	/*

		report, err := staticScan(&p)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Unable to scan")
		}

		logrus.WithFields(logrus.Fields{
			"job url": report.CeTaskURL,
		}).Info("Job url")

		task, err := waitForSonarJob(report)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Unable to get Job state")
		}

		status := getStatus(task, report)

		if status != p.Config.Quality {
			logrus.WithFields(logrus.Fields{
				"status": status,
			}).Fatal("QualityGate status failed")
		}

		return nil
	*/
}

func getStatus(task *TaskResponse, report *SonarReport) string {
	reportRequest := url.Values{
		"analysisId": {task.Task.AnalysisID},
	}
	projectRequest, err := http.NewRequest("GET", report.ServerURL+"/api/qualitygates/project_status?"+reportRequest.Encode(), nil)
	projectRequest.Header.Add("Authorization", "Basic "+os.Getenv("TOKEN"))
	projectResponse, err := netClient.Do(projectRequest)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed get status")
	}
	buf, _ := ioutil.ReadAll(projectResponse.Body)
	project := ProjectStatusResponse{}
	if err := json.Unmarshal(buf, &project); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed")
	}

	return project.ProjectStatus.Status
}

func getSonarJobStatus(report *SonarReport) *TaskResponse {

	taskRequest, err := http.NewRequest("GET", report.CeTaskURL, nil)
	taskRequest.Header.Add("Authorization", "Basic "+os.Getenv("TOKEN"))
	taskResponse, err := netClient.Do(taskRequest)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed get sonar job status")
	}
	buf, _ := ioutil.ReadAll(taskResponse.Body)
	task := TaskResponse{}
	json.Unmarshal(buf, &task)
	return &task
}

func waitForSonarJob(report *SonarReport) (*TaskResponse, error) {
	timeout := time.After(300 * time.Second)
	tick := time.Tick(500 * time.Millisecond)
	for {
		select {
		case <-timeout:
			return nil, errors.New("timed out")
		case <-tick:
			job := getSonarJobStatus(report)
			if job.Task.Status == "SUCCESS" {
				return job, nil
			}
		}
	}
}

func staticScan(p *Plugin) (*SonarReport, error) {
	if _, err := os.Stat(p.Config.Settings); errors.Is(err, os.ErrExist) {
		cmd := exec.Command("sed", "-e", "s/=/=\"/", "-e", "s/$/\"/", p.Config.Settings)
		output, err := cmd.CombinedOutput()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Run command sed failed")
			return nil, err
		}
		// log.Printf("%s\n",output)

		report := SonarReport{}
		err = toml.Unmarshal(output, &report)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Toml Unmarshal failed")
			return nil, err
		}

	}
	args := []string{
		"-Dsonar.projectKey=" + strings.Replace(p.Config.Key, "/", ":", -1),
		"-Dsonar.projectName=" + p.Config.Name,
		"-Dsonar.host.url=" + p.Config.Host,
		"-Dsonar.login=" + p.Config.Token,
		"-Dsonar.projectVersion=" + p.Config.Version,
		"-Dsonar.sources=" + p.Config.Sources,
		"-Dproject.settings=" + p.Config.Settings,
		"-Dsonar.ws.timeout=360",
		"-Dsonar.inclusions=" + p.Config.Inclusions,
		"-Dsonar.exclusions=" + p.Config.Exclusions,
		"-Dsonar.profile=" + p.Config.Profile,
		"-Dsonar.branch=" + p.Config.Branch,
		"-Dsonar.scm.provider=git",
		//"-Dsonar.java.binaries=" + p.Binaries,
	}

	cmd := exec.Command("sonar-scanner", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Run command sonar-scanner failed")
		return nil, err
	}
	fmt.Printf("out:\n%s", output)
	cmd = exec.Command("sed", "-e", "s/=/=\"/", "-e", "s/$/\"/", ".scannerwork/report-task.txt")
	output, err = cmd.CombinedOutput()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Run command sed failed")
		return nil, err
	}
	// log.Printf("%s\n",output)

	report := SonarReport{}
	err = toml.Unmarshal(output, &report)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Toml Unmarshal failed")
		return nil, err
	}

	return &report, nil
}
