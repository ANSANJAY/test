package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	workflowName   = "SonarQube_Build"
	repoOwner      = "amex-eng"
	repoName       = "REPO_NAME" // replace with your repo name
	baseURL        = "https://api.github.com"
)

type WorkflowRuns struct {
	WorkflowRuns []struct {
		ID         int    `json:"id"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
	} `json:"workflow_runs"`
}

type Jobs struct {
	Jobs []struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Conclusion string `json:"conclusion"`
	} `json:"jobs"`
}

func main() {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("GITHUB_TOKEN environment variable not set")
	}

	client := &http.Client{}

	// Step 1: Get workflow runs
	workflowRunsURL := fmt.Sprintf("%s/repos/%s/%s/actions/workflows/%s/runs", baseURL, repoOwner, repoName, workflowName)
	req, err := http.NewRequest("GET", workflowRunsURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", "Bearer "+githubToken)
	req.Header.Add("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var runs WorkflowRuns
	if err := json.Unmarshal(body, &runs); err != nil {
		log.Fatal(err)
	}

	var failedRunID int
	for _, run := range runs.WorkflowRuns {
		if run.Conclusion == "failure" {
			failedRunID = run.ID
			break
		}
	}

	if failedRunID == 0 {
		fmt.Println("No failed workflow runs found.")
		return
	}

	// Step 2: Get jobs for the failed workflow run
	jobsURL := fmt.Sprintf("%s/repos/%s/%s/actions/runs/%d/jobs", baseURL, repoOwner, repoName, failedRunID)
	req, err = http.NewRequest("GET", jobsURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", "Bearer "+githubToken)
	req.Header.Add("Accept", "application/vnd.github+json")

	resp, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var jobs Jobs
	if err := json.Unmarshal(body, &jobs); err != nil {
		log.Fatal(err)
	}

	var failedJobID int
	for _, job := range jobs.Jobs {
		if job.Conclusion == "failure" {
			failedJobID = job.ID
			break
		}
	}

	if failedJobID == 0 {
		fmt.Println("No failed jobs found in the workflow run.")
		return
	}

	// Step 3: Get logs for the failed job
	logsURL := fmt.Sprintf("%s/repos/%s/%s/actions/jobs/%d/logs", baseURL, repoOwner, repoName, failedJobID)
	req, err = http.NewRequest("GET", logsURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", "Bearer "+githubToken)
	req.Header.Add("Accept", "application/vnd.github+json")

	resp, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	logs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Error logs for the failed job:")
	fmt.Println(strings.TrimSpace(string(logs)))
}