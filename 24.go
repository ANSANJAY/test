package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Function to get contributors for a repo using gh CLI
func getContributors(repo string) ([]string, error) {
	// Execute gh CLI command
	cmd := exec.Command("gh", "api", fmt.Sprintf("/repos/%s/contributors", repo), "--jq", ".[].login")
	output, err := cmd.Output()
	if err != nil {
		// Capture stderr to print useful error information
		if exitErr, ok := err.(*exec.ExitError); ok {
			errorOutput := string(exitErr.Stderr)
			return nil, fmt.Errorf("error fetching contributors for %s: %v. Details: %s", repo, err, errorOutput)
		}
		return nil, fmt.Errorf("error fetching contributors for %s: %v", repo, err)
	}

	// Split the output by newlines to get each contributor
	contributors := strings.Split(strings.TrimSpace(string(output)), "\n")
	return contributors, nil
}

func main() {
	// Open the CSV file
	file, err := os.Open("repos.csv")
	if err != nil {
		log.Fatalf("Could not open CSV file: %v", err)
	}
	defer file.Close()

	// Read the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Could not read CSV file: %v", err)
	}

	// Iterate over each row in the CSV (assuming repo names are in the first column)
	for i, record := range records {
		// Skip header if needed
		if i == 0 {
			continue
		}

		repoName := strings.TrimSpace(record[0]) // Ensure no extra spaces or newlines

		// Check if the repo name is valid
		if repoName == "" {
			fmt.Println("Skipping empty repo name.")
			continue
		}

		// Prepend "amex-eng/" to the repo name
		repo := fmt.Sprintf("amex-eng/%s", repoName)

		// Get the contributors for the repo
		contributors, err := getContributors(repo)
		if err != nil {
			fmt.Printf("Error fetching contributors for %s: %v\n", repo, err)
			continue
		}

		// Print the contributors for each repo
		if len(contributors) > 0 {
			fmt.Printf("Contributors for %s: %s\n", repo, strings.Join(contributors, ", "))
		} else {
			fmt.Printf("No contributors found for %s\n", repo)
		}
	}
}
