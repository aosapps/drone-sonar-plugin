package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"

	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"encoding/xml"

	"bytes"
)

var netClient *http.Client

var projectKey = ""

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
	}
	// SonarReport it is the representation of .scannerwork/report-task.txt
	SonarReport struct {
		ProjectKey   string `toml:"projectKey"`
		ServerURL    string `toml:"serverUrl"`
		DashboardURL string `toml:"dashboardUrl"`
		CeTaskID     string `toml:"ceTaskId"`
		CeTaskURL    string `toml:"ceTaskUrl"`
	}
	Plugin struct {
		Config Config
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
	Project struct {
		ProjectStatus Status `json:"projectStatus"`
	}
	Status struct {
		Status            string      `json:"status"`
		IgnoredConditions bool        `json:"ignoredConditions"`
		Conditions        []Condition `json:"conditions"`
	}

	Condition struct {
		Status         string `json:"status"`
		MetricKey      string `json:"metricKey"`
		Comparator     string `json:"comparator"`
		PeriodIndex    int    `json:"periodIndex"`
		ErrorThreshold string `json:"errorThreshold"`
		ActualValue    string `json:"actualValue"`
	}

	Testsuites struct {
		XMLName   xml.Name    `xml:"testsuites"`
		Text      string      `xml:",chardata"`
		TestSuite []Testsuite `xml:"testsuite"`
	}
	Testsuite struct {
		Text     string     `xml:",chardata"`
		Package  string     `xml:"package,attr"`
		Time     int        `xml:"time,attr"`
		Tests    int        `xml:"tests,attr"`
		Errors   int        `xml:"errors,attr"`
		Name     string     `xml:"name,attr"`
		TestCase []Testcase `xml:"testcase"`
	}

	Testcase struct {
		Text      string   `xml:",chardata"`
		Time      int      `xml:"time,attr"`      // Actual Value Sonar
		Name      string   `xml:"name,attr"`      // Metric Key
		Classname string   `xml:"classname,attr"` // The metric Rule
		Failure   *Failure `xml:"failure"`        // Sonar Failure - show results
	}
	Failure struct {
		Text    string `xml:",chardata"`
		Message string `xml:"message,attr"`
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

func TryCatch(f func()) func() error {
	return func() (err error) {
		defer func() {
			if panicInfo := recover(); panicInfo != nil {
				err = fmt.Errorf("%v", panicInfo)
				return
			}
		}()
		f() // calling the decorated function
		return err
	}
}
func ParseJunit(data_arr Project, projectName string) Testsuites {
	errors := 0
	total := 0
	testCases := []Testcase{}

	conditionsArray := data_arr.ProjectStatus.Conditions

	for _, condition := range conditionsArray {
		total += 1
		if condition.Status != "OK" {
			errors += 1
			cond := &Testcase{
				Name: condition.MetricKey, Classname: "Violate if " + condition.ActualValue + " is " + condition.Comparator + " " + condition.ErrorThreshold, Failure: &Failure{Message: "Violated: " + condition.ActualValue + " is " + condition.Comparator + " " + condition.ErrorThreshold},
			}
			testCases = append(testCases, *cond)
		} else {
			cond := &Testcase{Name: condition.MetricKey, Classname: "Violate if " + condition.ActualValue + " is " + condition.Comparator + " " + condition.ErrorThreshold, Time: 0}
			testCases = append(testCases, *cond)
		}
	}
	SonarJunitReport := &Testsuites{
		TestSuite: []Testsuite{
			Testsuite{
				Time: 13, Package: projectName, Errors: errors, Tests: total, Name: "Harness CIE Sonar Test", TestCase: testCases,
			},
		},
	}

	out, _ := xml.MarshalIndent(SonarJunitReport, " ", "  ")
	fmt.Println(string(out))
	fmt.Printf("\n")
	out, _ = xml.MarshalIndent(testCases, " ", "  ")
	fmt.Println(string(out))
	fmt.Printf("\n")
	//SonarJunitReport.TestSuite.TestCase := TestCases

	return *SonarJunitReport
}

func GetProjectKey(key string) string {
	projectKey = strings.Replace(key, "/", ":", -1)
	return projectKey
}
func (p Plugin) Exec() error {
	args := []string{
		"-Dsonar.host.url=" + p.Config.Host,
		"-Dsonar.login=" + p.Config.Token,
	}
	//projectFinalKey := strings.Replace(p.Config.Key, "/", ":", -1)
	projectFinalKey := p.Config.Key

	if !p.Config.UsingProperties {
		argsParameter := []string{
			"-Dsonar.projectKey=" + projectFinalKey,
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
			"-Dsonar.qualitygate.wait=true",
		}
		args = append(args, argsParameter...)
	}

	if p.Config.BranchAnalysis {
		args = append(args, "-Dsonar.branch.name="+p.Config.Branch)
	}
	fmt.Printf("sonar-scanner")
	fmt.Printf("%v", args)
	cmd := exec.Command("sonar-scanner", args...)
	// fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("==> Code Analysis Result:\n")
	err := cmd.Run()
	if err != nil {
		return err
	}

	/* Test result report*/

	cmd = exec.Command("cat", ".scannerwork/report-task.txt")
	// JUNIT REPORT
	junitReport := ""
	// import "bytes"
	buf := new(bytes.Buffer)

	// JUNIT REPORT
	cmd.Stdout = buf

	cmd.Stderr = os.Stderr
	fmt.Printf("==> Report Result:\n")
	junitReport = buf.String() // returns a string of what was written to it
	fmt.Printf(junitReport)
	err = cmd.Run()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Run command cat reportname failed")
		return err
	}

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
		return err
	}

	status := getStatus(task, report)

	// JUNIT REPORT
	fmt.Printf("---> JUNIT Test <-------------------------------------------------\n")
	//readBuf, _ := ioutil.ReadAll(buf)
	bytesReport := []byte(junitReport)
	fmt.Printf("BEGIN")
	fmt.Printf(junitReport)
	fmt.Printf("END")
	var project Project
	err = json.Unmarshal(bytesReport, &project)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", project)
	fmt.Printf("---> JUNIT Test <-------------------------------------------------\n")
	// JUNIT

	if status != p.Config.Quality {
		logrus.WithFields(logrus.Fields{
			"status": status,
		}).Fatal("QualityGate status failed")
	} else {
		fmt.Printf("\n==> PASSED <==\n")
		logrus.WithFields(logrus.Fields{
			"status": status,
		}).Info("Quality Gate Status Success")
		fmt.Printf("\n")
		fmt.Printf("==> SONAR PROJECT DASHBOARD <==\n")
		fmt.Printf(p.Config.Host)
		fmt.Printf("/dashboard?id=")
		fmt.Printf(p.Config.Name)
		fmt.Printf("\n==> Harness CIE SonarQube Plugin with Quality Gateway <==\n\n")
	}

	return nil
}

func staticScan(p *Plugin) (*SonarReport, error) {

	/* end Test Report Result */
	//fmt.Printf("==> Sed Result:\n")
	cmd := exec.Command("sed", "-e", "s/=/=\"/", "-e", "s/$/\"/", ".scannerwork/report-task.txt")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Run command sed failed")
		return nil, err
	}
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
	fmt.Printf("==> Report Result:\n")
	fmt.Printf(string(buf))
	fmt.Printf("\n\n==> Harness CIE SonarQube Plugin with Quality Gateway <==\n\n==> Results:")

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
			if job.Task.Status == "ERROR" {
				return nil, errors.New("ERROR")
			}
		}
	}
}
