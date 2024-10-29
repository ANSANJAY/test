package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// PRStatus represents the structure for PR status response
type PRStatus struct {
	State string `json:"state"`
}

func getPRStatus(repoName, prNumber string) (string, error) {
	// Execute gh command to get the PR status
	cmd := exec.Command("gh", "pr", "view", prNumber, "--repo", "org-name/"+repoName, "--json", "state")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error fetching PR status: %v", err)
	}

	// Parse JSON to extract the "state" field
	var status PRStatus
	if err := json.Unmarshal(output, &status); err != nil {
		return "", fmt.Errorf("error parsing PR status JSON: %v", err)
	}

	return status.State, nil
}

func processCSV(inputFilePath, outputFilePath string) error {
	// Open the input CSV file
	file, err := os.Open(inputFilePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading CSV file: %v", err)
	}

	// Create the output CSV file
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output CSV file: %v", err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write the header row to the output file
	writer.Write([]string{"Repo Name", "Status"})

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

		// Write the result to the output CSV with only the repo name and status
		writer.Write([]string{repoName, prStatus})
	}

	return nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <input_file.csv> <output_file.csv>")
		return
	}
	inputFilePath := os.Args[1]
	outputFilePath := os.Args[2]
	if err := processCSV(inputFilePath, outputFilePath); err != nil {
		fmt.Printf("Error processing CSV: %v\n", err)
	}
}