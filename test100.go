package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

const (
	repo = "amex-eng/REPO_NAME" // replace with your repo name
)

// WorkflowRun represents the workflow run details
type WorkflowRun struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

func main() {
	// Step 1: Get recent workflow runs and their statuses
	cmd := exec.Command("gh", "api", fmt.Sprintf("repos/%s/actions/runs", repo), "--jq", ".workflow_runs[] | {id: .id, name: .name, status: .status, conclusion: .conclusion}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to fetch workflow runs: %v", err)
	}

	// Parse the output into a slice of WorkflowRun structs
	var runs []WorkflowRun
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		var run WorkflowRun
		if err := json.Unmarshal([]byte(line), &run); err == nil {
			runs = append(runs, run)
		}
	}

	// Print the names and statuses of the workflows
	fmt.Println("Workflow Name and Status:")
	for _, run := range runs {
		fmt.Printf("Name: %s, Status: %s, Conclusion: %s\n", run.Name, run.Status, run.Conclusion)
	}

	// Step 2: Find the latest failed workflow run
	var failedRunID int
	for _, run := range runs {
		if run.Conclusion == "failure" {
			failedRunID = run.ID
			break
		}
	}

	if failedRunID == 0 {
		fmt.Println("No failed workflow runs found.")
		return
	}

	// Step 3: Get jobs for the latest failed workflow run
	cmd = exec.Command("gh", "api", fmt.Sprintf("repos/%s/actions/runs/%d/jobs", repo, failedRunID), "--jq", ".jobs | .[] | select(.conclusion == \"failure\") | .id")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to fetch jobs for the workflow run: %v", err)
	}

	failedJobID := strings.TrimSpace(string(output))
	if failedJobID == "" {
		fmt.Println("No failed jobs found in the workflow run.")
		return
	}

	// Step 4: Get logs for the failed job
	cmd = exec.Command("gh", "api", fmt.Sprintf("repos/%s/actions/jobs/%s/logs", repo, failedJobID))
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to fetch logs for the failed job: %v", err)
	}

	fmt.Println("\nError logs for the failed job:")
	fmt.Println(strings.TrimSpace(string(output)))
}