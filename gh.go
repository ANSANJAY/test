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

// Worker function to process each repo
func processRepo(repo string, wg *sync.WaitGroup, semaphore chan struct{}) {
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

	if status == "completed" {
		fmt.Printf("The build is complete for repo %s/%s.\n", owner, repo)
	} else {
		fmt.Printf("The build is not complete yet for repo %s/%s.\n", owner, repo)
	}

	// Check for open pull requests
	prs, err := getOpenPullRequests(repo)
	if err != nil {
		fmt.Printf("Error fetching pull requests for %s/%s: %v\n", owner, repo, err)
		return
	}

	if len(prs) > 0 {
		fmt.Printf("Open Pull Requests for repo %s/%s:\n", owner, repo)
		for _, pr := range prs {
			if pr != "" {
				fmt.Println("- " + pr)
			}
		}
	} else {
		fmt.Printf("No open pull requests found for repo %s/%s.\n", owner, repo)
	}
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

func main() {
	// Read repository names from CSV file
	repos, err := readReposFromCSV(csvFilePath)
	if err != nil {
		fmt.Printf("Error reading CSV file: %v\n", err)
		return
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrentRequests) // Semaphore to limit concurrency

	start := time.Now()

	for _, repo := range repos {
		wg.Add(1)
		go processRepo(repo, &wg, semaphore)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	fmt.Printf("Processed %d repositories in %v\n", len(repos), time.Since(start))
}
