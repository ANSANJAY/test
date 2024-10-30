package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Function to get reviewers for a PR using gh CLI
func getReviewers(repo string, prNumber string) ([]string, error) {
	// Execute gh CLI command to fetch reviewers
	cmd := exec.Command("gh", "api", fmt.Sprintf("/repos/%s/pulls/%s/reviews", repo, prNumber), "--jq", ".[].user.login")
	output, err := cmd.Output()
	if err != nil {
		// Capture stderr for detailed error information
		if exitErr, ok := err.(*exec.ExitError); ok {
			errorOutput := string(exitErr.Stderr)
			return nil, fmt.Errorf("error fetching reviewers for %s PR #%s: %v. Details: %s", repo, prNumber, err, errorOutput)
		}
		return nil, fmt.Errorf("error fetching reviewers for %s PR #%s: %v", repo, prNumber, err)
	}

	// Split the output by newlines to get each reviewer
	reviewers := strings.Split(strings.TrimSpace(string(output)), "\n")
	return reviewers, nil
}

func main() {
	// Open the CSV file containing the list of PR links
	file, err := os.Open("prs.csv")
	if err != nil {
		log.Fatalf("Could not open CSV file: %v", err)
	}
	defer file.Close()

	// Create a new CSV file to store the output
	outputFile, err := os.Create("reviewers_output.csv")
	if err != nil {
		log.Fatalf("Could not create output CSV file: %v", err)
	}
	defer outputFile.Close()

	// Initialize the CSV writer
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write the CSV header
	writer.Write([]string{"Repo Name", "PR Link", "Reviewer Names"})

	// Read the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Could not read CSV file: %v", err)
	}

	// Iterate over each row in the CSV (assuming PR links are in the first column)
	for i, record := range records {
		// Skip header if needed
		if i == 0 {
			continue
		}

		prLink := strings.TrimSpace(record[0])
		parts := strings.Split(prLink, "/")
		if len(parts) < 7 {
			fmt.Printf("Invalid PR link format: %s\n", prLink)
			continue
		}

		// Extract repo name and PR number
		repo := fmt.Sprintf("abc/%s", parts[4])
		prNumber := parts[6]

		// Get the reviewers for the PR
		reviewers, err := getReviewers(repo, prNumber)
		if err != nil {
			fmt.Printf("Error fetching reviewers for %s PR #%s: %v\n", repo, prNumber, err)
			continue
		}

		// Write to the output CSV
		reviewersString := strings.Join(reviewers, ", ")
		writer.Write([]string{parts[4], prLink, reviewersString})
	}

	fmt.Println("CSV file 'reviewers_output.csv' has been created with PR links and reviewer names.")
}
