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
	// Execute gh CLI command to fetch contributors
	cmd := exec.Command("gh", "api", fmt.Sprintf("/repos/%s/contributors", repo), "--jq", ".[].login")
	output, err := cmd.Output()
	if err != nil {
		// Capture stderr for detailed error information
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

// Function to get the email of a contributor from their commits
func getContributorEmail(repo, contributor string) (string, error) {
	// Execute gh CLI command to fetch the contributor's latest commits and extract email
	cmd := exec.Command("gh", "api", fmt.Sprintf("/repos/%s/commits?author=%s", repo, contributor), "--jq", ".[0].commit.author.email")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			errorOutput := string(exitErr.Stderr)
			return "", fmt.Errorf("error fetching email for %s in repo %s: %v. Details: %s", contributor, repo, err, errorOutput)
		}
		return "", fmt.Errorf("error fetching email for %s in repo %s: %v", contributor, repo, err)
	}

	// Return the email address
	email := strings.TrimSpace(string(output))
	if email == "" {
		return "No email found", nil
	}
	return email, nil
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

		// For each contributor, get their email
		for _, contributor := range contributors {
			email, err := getContributorEmail(repo, contributor)
			if err != nil {
				fmt.Printf("Error fetching email for %s in repo %s: %v\n", contributor, repo, err)
				continue
			}

			fmt.Printf("Contributor: %s, Email: %s\n", contributor, email)
		}
	}
}
