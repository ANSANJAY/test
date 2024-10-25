package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

const (
	workflowName = "SonarQube_Build"
	repo         = "amex-eng/REPO_NAME" // replace with your repo name
)

func main() {
	// Step 1: Get the latest workflow run with the specific name
	cmd := exec.Command("gh", "api", fmt.Sprintf("repos/%s/actions/workflows/%s/runs", repo, workflowName), "--jq", ".workflow_runs | sort_by(.created_at) | reverse | .[0] | select(.conclusion == \"failure\") | .id")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to fetch workflow runs: %v", err)
	}

	failedRunID := strings.TrimSpace(string(output))
	if failedRunID == "" {
		fmt.Println("No failed workflow runs found.")
		return
	}

	// Step 2: Get jobs for the latest failed workflow run
	cmd = exec.Command("gh", "api", fmt.Sprintf("repos/%s/actions/runs/%s/jobs", repo, failedRunID), "--jq", ".jobs | .[] | select(.conclusion == \"failure\") | .id")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to fetch jobs for the workflow run: %v", err)
	}

	failedJobID := strings.TrimSpace(string(output))
	if failedJobID == "" {
		fmt.Println("No failed jobs found in the workflow run.")
		return
	}

	// Step 3: Get logs for the failed job
	cmd = exec.Command("gh", "api", fmt.Sprintf("repos/%s/actions/jobs/%s/logs", repo, failedJobID))
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to fetch logs for the failed job: %v", err)
	}

	fmt.Println("Error logs for the latest failed job:")
	fmt.Println(strings.TrimSpace(string(output)))
}