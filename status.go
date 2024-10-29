package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// getPRStatus uses gh CLI to get the status of a pull request
func getPRStatus(repo, prNumber string) (string, error) {
	// Construct the gh command to get the PR state
	cmd := exec.Command("gh", "pr", "view", fmt.Sprintf("%s#%s", repo, prNumber), "--json", "state")

	// Execute the command and capture output
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error fetching PR status for %s#%s: %v", repo, prNumber, err)
	}

	// Extract the state from the output
	state := strings.TrimSuffix(strings.Split(string(output), ":")[1], "\"}\n")
	state = strings.Trim(state, " \"")
	return state, nil
}

func main() {
	inputFile := "pull_requests.csv"
	outputFile := "pr_status_results.csv"

	// Open input CSV file
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		return
	}
	defer file.Close()

	// Prepare output CSV file
	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Write header to output file
	writer.Write([]string{"Pull Request Link", "Status"})

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		prLink := scanner.Text()

		// Parse repo and PR number from the link
		parts := strings.Split(prLink, "/pull/")
		if len(parts) < 2 {
			fmt.Printf("Invalid PR link format: %s\n", prLink)
			continue
		}
		repo := strings.Split(parts[0], "github.com/")[1]
		prNumber := parts[1]

		// Fetch PR status
		status, err := getPRStatus(repo, prNumber)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Write result to output CSV
		writer.Write([]string{prLink, status})
		fmt.Printf("PR %s status: %s\n", prLink, status)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
	}
}