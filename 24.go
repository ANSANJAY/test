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

		repo := record[0] // Assuming the repo name is in the first column
		contributors, err := getContributors(repo)
		if err != nil {
			fmt.Printf("Error fetching contributors for %s: %v\n", repo, err)
			continue
		}

		// Print the contributors for each repo
		fmt.Printf("Contributors for %s: %s\n", repo, strings.Join(contributors, ", "))
	}
}
