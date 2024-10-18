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

// Function to check the latest workflow run status using gh CLI for a specific repo
func getWorkflowStatus(repo string) (string, error) {
	cmd := exec.Command("gh", "run", "list", "--repo", fmt.Sprintf("%s/%s", owner, repo), "--workflow", workflowName, "--limit", "1", "--json", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command: %v", err)
	}

	// Check if the output is empty, indicating no runs were found
	if strings.Contains(string(output), "[]") {
		return "build not triggered", nil
	}

	// Check for specific statuses in the JSON output
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
	err = writer.Write([]string{"Repo Name", "Build Status"})
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

			// Fetch workflow status
			status, err := getWorkflowStatus(repo)
			if err != nil {
				fmt.Printf("Error fetching status for repo %s: %v\n", repo, err)
				status = "error"
			}

			// Collect result
			results[i] = []string{repo, status}
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
