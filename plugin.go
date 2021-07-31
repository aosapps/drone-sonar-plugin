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
	// Plugin defines the sonar-scaner plugin parameters.
	Plugin struct {
		Host       string
		Token      string
		Key        string
		Name       string
		Version    string
		Sources    string
		Inclusions string
		Exclusions string
		Language   string
		Profile    string
		Encoding   string
		Remote     string
		Branch     string
		Quality    string
		Settings   string
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

// Exec executes the plugin step
func (p Plugin) Exec() error {

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

	if status != p.Quality {
		logrus.WithFields(logrus.Fields{
			"status": status,
		}).Fatal("QualityGate status failed")
	}

	return nil
}

func staticScan(p *Plugin) (*SonarReport, error) {
	if _, err := os.Stat(p.Settings); errors.Is(err, os.ErrExist) {
		cmd := exec.Command("sed", "-e", "s/=/=\"/", "-e", "s/$/\"/", p.Settings)
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
		"-Dsonar.projectKey=" + strings.Replace(p.Key, "/", ":", -1),
		"-Dsonar.projectName=" + p.Name,
		"-Dsonar.host.url=" + p.Host,
		"-Dsonar.login=" + p.Token,
		"-Dsonar.projectVersion=" + p.Version,
		"-Dsonar.sources=" + p.Sources,
		"-Dproject.settings=" + p.Settings,
		"-Dsonar.ws.timeout=360",
		"-Dsonar.inclusions=" + p.Inclusions,
		"-Dsonar.exclusions=" + p.Exclusions,
		"-Dsonar.profile=" + p.Profile,
		"-Dsonar.branch=" + p.Branch,
		"-Dsonar.scm.provider=git",
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
