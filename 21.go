package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	owner             = "amex-eng"            // Default owner for all repos
	inputCSVFilePath  = "repos.csv"           // Path to input CSV
	outputCSVFilePath = "output_pr_age.csv"   // Path to output CSV
	prTitle           = "this is a pr"        // The title of the PR we are checking
	maxConcurrency    = 5                     // Limit concurrency to avoid GitHub rate limits
)

// PR struct to capture JSON response from GitHub CLI
type PR struct {
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

// Function to get the age of a PR in days
func getPRAge(repo string) (string, error) {
	// Get the list of PRs in JSON format
	cmd := exec.Command("gh", "pr", "list", "--repo", fmt.Sprintf("%s/%s", owner, repo), "--state", "open", "--json", "title,created_at")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command: %v", err)
	}

	// Check if no PRs are found
	if strings.Contains(string(output), "[]") {
		return "no PR", nil
	}

	// Parse the JSON output
	var prs []PR
	if err := json.Unmarshal(output, &prs); err != nil {
		return "", fmt.Errorf("error parsing JSON output: %v", err)
	}

	// Search for the PR with the specified title
	for _, pr := range prs {
		if pr.Title == prTitle {
			createdAt, err := time.Parse(time.RFC3339, pr.CreatedAt)
			if err != nil {
				return "", fmt.Errorf("error parsing PR creation date: %v", err)
			}
			// Calculate the age of the PR in days
			age := time.Since(createdAt).Hours() / 24
			return fmt.Sprintf("%.0f days", age), nil
		}
	}

	return "PR not found", nil
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
	err = writer.Write([]string{"Repo Name", "PR Age"})
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

			// Get PR age
			prAge, err := getPRAge(repo)
			if err != nil {
				fmt.Printf("Error fetching PR age for repo %s: %v\n", repo, err)
				prAge = "error"
			}

			// Collect result
			results[i] = []string{repo, prAge}
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