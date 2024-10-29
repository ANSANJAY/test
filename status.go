package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func getPRStatus(repoName, prNumber string) (string, error) {
	// Execute gh command to get the PR status
	cmd := exec.Command("gh", "pr", "view", prNumber, "--repo", "org-name/"+repoName, "--json", "state")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error fetching PR status: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func processCSV(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading CSV file: %v", err)
	}

	// Iterate through the CSV records
	for i, row := range records {
		// Skip header if present
		if i == 0 && strings.Contains(strings.ToLower(row[0]), "pull_request_link") {
			continue
		}

		// Extract repo name and PR number from the pull request link
		linkParts := strings.Split(row[0], "/")
		if len(linkParts) < 4 {
			fmt.Printf("Invalid pull request link: %s\n", row[0])
			continue
		}
		repoName := linkParts[len(linkParts)-3]
		prNumber := linkParts[len(linkParts)-1]

		// Get the PR status
		prStatus, err := getPRStatus(repoName, prNumber)
		if err != nil {
			fmt.Printf("Repo: %s, PR #%s - Error: %v\n", repoName, prNumber, err)
			continue
		}
		fmt.Printf("Repo: %s, PR #%s - Status: %s\n", repoName, prNumber, prStatus)
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <path/to/your_file.csv>")
		return
	}
	filePath := os.Args[1]
	if err := processCSV(filePath); err != nil {
		fmt.Printf("Error processing CSV: %v\n", err)
	}
}