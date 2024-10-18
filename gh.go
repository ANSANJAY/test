package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// Function to get the status of a workflow by name
func getWorkflowStatusByName(owner, repo, workflowName string) (string, error) {
	cmd := exec.Command("gh", "run", "list", "--repo", fmt.Sprintf("%s/%s", owner, repo), "--workflow", workflowName, "--limit", "1", "--json", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Extracting status from the output
	outputStr := string(output)
	if strings.Contains(outputStr, `"status":"completed"`) {
		return "completed", nil
	} else if strings.Contains(outputStr, `"status":"in_progress"`) {
		return "in_progress", nil
	} else if strings.Contains(outputStr, `"status":"failure"`) {
		return "failure", nil
	}

	return "unknown", nil
}

func main() {
	owner := "amex-eng"
	repo := "ace-framework_ace-core-services"
	workflowName := "Automated SonarQube Integration"

	status, err := getWorkflowStatusByName(owner, repo, workflowName)
	if err != nil {
		fmt.Println("Error fetching workflow status:", err)
		return
	}

	fmt.Printf("Workflow '%s' in repo %s/%s is: %s\n", workflowName, owner, repo, status)
}
