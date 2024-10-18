package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	maxConcurrentRequests = 10 // Limit the number of concurrent API calls
	owner                 = "amex-eng"
	csvFilePath           = "repos.csv"
	outputCSVFilePath     = "output.csv"
)

// Function to check the latest workflow run status using gh CLI for a specific repo
func getLatestWorkflowStatus(repo string) (string, error) {
	cmd := exec.Command("gh", "run", "list", "--repo", fmt.Sprintf("%s/%s", owner, repo), "--limit", "1", "--json", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Extracting status
	outputStr := string(output)
	if strings.Contains(outputStr, `"status":"completed"`) {
		return "completed", nil
	}

	return "not completed", nil
}

// Function to check for open pull requests using gh CLI for a specific repo
func getOpenPullRequests(repo string) ([]string, error) {
	cmd := exec.Command("gh", "pr", "list", "--repo", fmt.Sprintf("%s/%s", owner, repo), "--json", "title")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	// Extracting PR titles
	outputStr := string(output)
	prs := strings.Split(outputStr, "\n")

	return prs, nil
}

// Function to process each repo and return the result
func processRepo(repo string, wg *sync.WaitGroup, semaphore chan struct{}, resultChan chan<- []string) {
	defer wg.Done()

	// Acquire semaphore to limit concurrency
	semaphore <- struct{}{}
	defer func() { <-semaphore }() // Release semaphore

	// Check workflow status
	status, err := getLatestWorkflowStatus(repo)
	if err != nil {
		fmt.Printf("Error fetching workflow status for %s/%s: %v\n", owner, repo, err)
		return
	}

	// Check for open pull requests
	prs, err := getOpenPullRequests(repo)
	if err != nil {
		fmt.Printf("Error fetching pull requests for %s/%s: %v\n", owner, repo, err)
		return
	}

	var prStatus string
	var prNames string
	if len(prs) > 0 && prs[0] != "" {
		prStatus = "yes"
		prNames = strings.Join(prs, ", ")
	} else {
		prStatus = "no"
		prNames = ""
	}

	// Send results to the result channel
	resultChan <- []string{repo, status, prStatus, prNames}
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
	for _, record := range records[1:] { // Skip header
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
	writer.Write([]string{"Repo Name", "Build Status", "PR Raised", "Name of PR"})

	// Write each record
	for _, record := range results {
		writer.Write(record)
	}

	return nil
}

func main() {
	// Read repository names from CSV file
	repos, err := readReposFromCSV(csvFilePath)
	if err != nil {
		fmt.Printf("Error reading CSV file: %v\n", err)
		return
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrentRequests) // Semaphore to limit concurrency
	resultChan := make(chan []string, len(repos))           // Channel to collect results

	start := time.Now()

	for _, repo := range repos {
		wg.Add(1)
		go processRepo(repo, &wg, semaphore, resultChan)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultChan)

	// Collect results from the channel
	var results [][]string
	for result := range resultChan {
		results = append(results, result)
	}

	// Write results to the output CSV file
	err = writeResultsToCSV(outputCSVFilePath, results)
	if err != nil {
		fmt.Printf("Error writing results to CSV: %v\n", err)
		return
	}

	fmt.Printf("Processed %d repositories and saved results to %s in %v\n", len(repos), outputCSVFilePath, time.Since(start))
}
