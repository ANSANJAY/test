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

	// Write the header row to the output file (adding PR Status column)
	header := append(records[0], "PR Status")
	writer.Write(header)

	// Cache to store previously fetched PR statuses
	prStatusCache := make(map[string]string)

	// Iterate through the CSV records, starting from the second row (skipping header)
	for _, row := range records[1:] {
		pullRequestLink := row[0]
		repoName := row[1]
		prNumber := strings.Split(pullRequestLink, "/")[len(strings.Split(pullRequestLink, "/"))-1]

		// Check if we already fetched the status for this PR link to avoid redundant API calls
		var prStatus string
		var err error
		if cachedStatus, exists := prStatusCache[pullRequestLink]; exists {
			prStatus = cachedStatus
		} else {
			prStatus, err = getPRStatus(repoName, prNumber)
			if err != nil {
				fmt.Printf("Error fetching status for %s: %v\n", pullRequestLink, err)
				prStatus = "Error"
			}
			// Cache the status to reuse for duplicate PR links
			prStatusCache[pullRequestLink] = prStatus
		}

		// Append PR status to the row and write to the output CSV
		outputRow := append(row, prStatus)
		writer.Write(outputRow)
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