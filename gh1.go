package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const (
	owner             = "amex-eng"            // Default owner for all repos
	inputCSVFilePath  = "repos.csv"           // Path to input CSV
	outputCSVFilePath = "output_status.csv"   // Path to output CSV
	workflowName      = "SonarQube_Build"     // The correct workflow name for checking status
	maxConcurrency    = 5                     // Limit concurrency to avoid GitHub rate limits
	prTitle           = "this is a pr"        // The title of the PR we are checking
)

// Function to check the latest workflow run status and conclusion using gh CLI for a specific repo
func getWorkflowStatus(repo string) (string, string, error) {
	cmd := exec.Command("gh", "run", "list", "--repo", fmt.Sprintf("%s/%s", owner, repo), "--workflow", workflowName, "--limit", "1", "--json", "status,conclusion")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("error running command: %v", err)
	}

	// Check if the output is empty, indicating no runs were found
	if strings.Contains(string(output), "[]") {
		return "build not triggered", "N/A", nil
	}

	// Check for specific statuses in the JSON output
	outputStr := string(output)
	status := "unknown"
	conclusion := "N/A"

	if strings.Contains(outputStr, `"status":"completed"`) {
		status = "completed"
		if strings.Contains(outputStr, `"conclusion":"success"`) {
			conclusion = "success"
		} else if strings.Contains(outputStr, `"conclusion":"failure"`) {
			conclusion = "failure"
		} else {
			// If it's neither success nor failure, check for other conclusions
			conclusion = extractConclusion(outputStr)
		}
	} else if strings.Contains(outputStr, `"status":"in_progress"`) {
		status = "in_progress"
		conclusion = "N/A"
	} else if strings.Contains(outputStr, `"status":"failure"`) {
		status = "failure"
		conclusion = "N/A"
	}

	return status, conclusion, nil
}

// Helper function to extract the conclusion from the JSON output
func extractConclusion(output string) string {
	possibleConclusions := []string{"neutral", "cancelled", "timed_out", "action_required", "skipped"}
	for _, conclusion := range possibleConclusions {
		if strings.Contains(output, `"conclusion":"`+conclusion+`"`) {
			return conclusion
		}
	}
	return "unknown"
}

// Function to check if a PR with a specific title exists for the repo
func checkIfPRExists(repo string) (bool, error) {
	cmd := exec.Command("gh", "pr", "list", "--repo", fmt.Sprintf("%s/%s", owner, repo), "--state", "open", "--json", "title")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("error running command: %v", err)
	}

	// Check if the output contains the specific PR title
	outputStr := string(output)
	if strings.Contains(outputStr, prTitle) {
		return true, nil
	}

	return false, nil
}

// Function to read repository names from a CSV file
func readReposFromCSV(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var repos []string
	for i, record := range records {
		// Skip header
		if i == 0 {
			continue
		}
		repos = append(repos, record[0])
	}

	return repos, nil
}

// Function to write results to a CSV file
func writeResultsToCSV(filePath string, results [][]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	err = writer.Write([]string{"Repo Name", "Build Status", "Conclusion", "PR Raised"})
	if err != nil {
		return err
	}

	// Write each record
	for _, record := range results {
		err = writer.Write(record)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// Read repository names from CSV file
	repos, err := readReposFromCSV(inputCSVFilePath)
	if err != nil {
		fmt.Printf("Error reading CSV file: %v\n", err)
		return
	}

	// Channel to limit concurrent API calls
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	results := make([][]string, len(repos))

	// Process each repository concurrently
	for i, repo := range repos {
		wg.Add(1)
		go func(i int, repo string) {
			defer wg.Done()

			semaphore <- struct{}{} // Acquire slot in semaphore
			defer func() { <-semaphore }() // Release slot

			// Fetch workflow status and conclusion
			status, conclusion, err := getWorkflowStatus(repo)
			if err != nil {
				fmt.Printf("Error fetching status for repo %s: %v\n", repo, err)
				status = "error"
				conclusion = "N/A"
			}

			// Check if PR with specific title is raised if the build is completed
			prRaised := "N/A"
			if status == "completed" {
				prExists, err := checkIfPRExists(repo)
				if err != nil {
					fmt.Printf("Error checking PR for repo %s: %v\n", repo, err)
					prRaised = "error"
				} else if prExists {
					prRaised = "yes"
				} else {
					prRaised = "no"
				}
			}

			// Collect result
			results[i] = []string{repo, status, conclusion, prRaised}
		}(i, repo)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Write results to the output CSV
	err = writeResultsToCSV(outputCSVFilePath, results)
	if err != nil {
		fmt.Printf("Error writing results to CSV: %v\n", err)
		return
	}

	fmt.Printf("Processed %d repositories and saved results to %s\n", len(repos), outputCSVFilePath)
}
