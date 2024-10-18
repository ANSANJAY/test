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
)

// Function to check the latest workflow run status and conclusion using gh CLI for a specific repo
func getWorkflowStatus(repo string) (string, string, string, error) {
	cmd := exec.Command("gh", "run", "list", "--repo", fmt.Sprintf("%s/%s", owner, repo), "--workflow", workflowName, "--limit", "1", "--json", "status,conclusion,head_sha")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", "", fmt.Errorf("error running command: %v", err)
	}

	// Check if the output is empty, indicating no runs were found
	if strings.Contains(string(output), "[]") {
		return "build not triggered", "", "", nil
	}

	// Check for specific statuses and conclusions in the JSON output
	outputStr := string(output)
	var status, conclusion, sha string

	if strings.Contains(outputStr, `"status":"completed"`) {
		status = "completed"
		if strings.Contains(outputStr, `"conclusion":"success"`) {
			conclusion = "success"
		} else if strings.Contains(outputStr, `"conclusion":"failure"`) {
			conclusion = "failure"
		} else {
			conclusion = "unknown"
		}
		// Extract SHA to reference the run if we need to retrieve detailed logs
		shaStart := strings.Index(outputStr, `"head_sha":"`) + len(`"head_sha":"`)
		shaEnd := strings.Index(outputStr[shaStart:], `"`)
		sha = outputStr[shaStart : shaStart+shaEnd]
	} else if strings.Contains(outputStr, `"status":"in_progress"`) {
		status = "in_progress"
		conclusion = "N/A"
	} else {
		status = "unknown"
		conclusion = "N/A"
	}

	return status, conclusion, sha, nil
}

// Function to retrieve failure details for a specific run (if any)
func getErrorDetails(repo, sha string) (string, error) {
	if sha == "" {
		return "", fmt.Errorf("SHA is empty, cannot retrieve logs")
	}

	cmd := exec.Command("gh", "run", "view", sha, "--repo", fmt.Sprintf("%s/%s", owner, repo), "--log")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error fetching error details: %v", err)
	}

	// Extract error details from the logs
	outputStr := string(output)
	if strings.Contains(outputStr, "failed") {
		// Extract some relevant lines that indicate the failure
		return outputStr, nil
	}

	return "no error details found", nil
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
	err = writer.Write([]string{"Repo Name", "Build Status", "Build Conclusion", "Error Details"})
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

			// Fetch workflow status, conclusion, and SHA for the latest run
			status, conclusion, sha, err := getWorkflowStatus(repo)
			var errorDetails string
			if err != nil {
				fmt.Printf("Error fetching status for repo %s: %v\n", repo, err)
				status = "error"
				conclusion = "N/A"
				errorDetails = "N/A"
			} else {
				// Only try fetching logs if the build failed
				if conclusion == "failure" && sha != "" {
					errorDetails, err = getErrorDetails(repo, sha)
					if err != nil {
						errorDetails = fmt.Sprintf("error retrieving details: %v", err)
					}
				} else {
					errorDetails = "N/A"
				}
			}

			// Collect result
			results[i] = []string{repo, status, conclusion, errorDetails}
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
