package main

import (
	"fmt"
	sonargo "github.com/magicsong/sonargo/sonar"
	"regexp"
	"time"
)

const SONAR_TIME_FORMAT = "2006-01-02T15:04:05-0700"

// How long to wait between consecutive API requests (seconds)
const SONAR_REQUEST_SLEEP = 10

/**
 * Extracts the report id from code analysis result
 * The corresponding line looks like this:
	  More about the report processing at https://sonarcloud.io/api/ce/task?id=AW6aSZwxGXJ8Zd7jkmP2
*/
func (p Plugin) extractReportIdFromAnalysisLog(analysisLog string) (string, error) {
	var re = regexp.MustCompile(`(?m)More about the report processing at .*\/api\/ce\/task\?id=(.*)$`)
	matches := re.FindStringSubmatch(analysisLog)
	if len(matches) != 2 { // expect exactly 2 results. 0 is full string. 1 is the id
		return "", fmt.Errorf("unable to get report processing url from analysis result.")
	}
	return matches[1], nil
}

func (p Plugin) getCompletedTaskReport(taskId string) (*sonargo.CeTaskObject, error) {
	for {
		taskObject, _, apiError := p.SonarClient.Ce.Task(&sonargo.CeTaskOption{Id: taskId})
		if apiError != nil {
			return nil, apiError
		}
		startedTime, timeErr := time.Parse(SONAR_TIME_FORMAT, taskObject.Task.StartedAt)
		if timeErr != nil {
			return nil, timeErr
		}
		taskStatus := taskObject.Task.Status
		if taskStatus != "SUCCESS" && taskStatus != "FAILED" {
			fmt.Printf("Awaiting completion of analysis. Current status is \"%s\". Analysis started on %s.\n", taskStatus, startedTime)
			time.Sleep(SONAR_REQUEST_SLEEP * time.Second)
			continue
		}
		completedTime, timeErr := time.Parse(SONAR_TIME_FORMAT, taskObject.Task.ExecutedAt)
		if timeErr != nil {
			return nil, timeErr
		}
		fmt.Printf("Analysis completed on %s with status \"%s\".\n", completedTime, taskStatus)
		if taskStatus == "FAILED" {
			return nil, fmt.Errorf("pipeline aborted because processing by the Sonar server failed")
		}
		return taskObject, nil
	}
}

func (p Plugin) validateQualityGate(taskId string) error {
	// Get completed analysis report
	ceTask, err := p.getCompletedTaskReport(taskId)
	if err != nil {
		return err
	}

	// Check Quality Gate
	projectStatusOption := &sonargo.QualitygatesProjectStatusOption{AnalysisId: ceTask.Task.AnalysisID}
	qualitygate, _, err := p.SonarClient.Qualitygates.ProjectStatus(projectStatusOption)
	if err != nil {
		return err
	}
	if qualitygate.ProjectStatus.Status == "ERROR" {
		return fmt.Errorf("pipeline aborted because quality gate failed")
	}
	return nil
}
