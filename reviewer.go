package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Function to get reviewers for a PR using gh CLI
func getReviewers(repo string, prNumber string) ([]string, error) {
	cmd := exec.Command("gh", "api", fmt.Sprintf("/repos/%s/pulls/%s/reviews", repo, prNumber), "--jq", ".[].user.login")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			errorOutput := string(exitErr.Stderr)
			return nil, fmt.Errorf("error fetching reviewers for %s PR #%s: %v. Details: %s", repo, prNumber, err, errorOutput)
		}
		return nil, fmt.Errorf("error fetching reviewers for %s PR #%s: %v", repo, prNumber, err)
	}

	reviewers := strings.Split(strings.TrimSpace(string(output)), "\n")
	return reviewers, nil
}

// Function to get the email of a reviewer from their latest commit in the PR
func getReviewerEmail(repo, contributor string) (string, error) {
	cmd := exec.Command("gh", "api", fmt.Sprintf("/repos/%s/commits?author=%s", repo, contributor), "--jq", ".[0].commit.author.email")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			errorOutput := string(exitErr.Stderr)
			return "", fmt.Errorf("error fetching email for %s in repo %s: %v. Details: %s", contributor, repo, err, errorOutput)
		}
		return "", fmt.Errorf("error fetching email for %s in repo %s: %v", contributor, repo, err)
	}

	email := strings.TrimSpace(string(output))
	if email == "" {
		return "No email found", nil
	}
	return email, nil
}

func main() {
	file, err := os.Open("prs.csv")
	if err != nil {
		log.Fatalf("Could not open CSV file: %v", err)
	}
	defer file.Close()

	outputFile, err := os.Create("reviewers_with_emails.csv")
	if err != nil {
		log.Fatalf("Could not create output CSV file: %v", err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	writer.Write([]string{"Repo Name", "PR Link", "Reviewer Name", "Reviewer Email"})

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Could not read CSV file: %v", err)
	}

	for i, record := range records {
		if i == 0 {
			continue
		}

		prLink := strings.TrimSpace(record[0])
		parts := strings.Split(prLink, "/")
		if len(parts) < 7 {
			fmt.Printf("Invalid PR link format: %s\n", prLink)
			continue
		}

		repo := fmt.Sprintf("abc/%s", parts[4])
		prNumber := parts[6]

		reviewers, err := getReviewers(repo, prNumber)
		if err != nil {
			fmt.Printf("Error fetching reviewers for %s PR #%s: %v\n", repo, prNumber, err)
			continue
		}

		for _, reviewer := range reviewers {
			email, err := getReviewerEmail(repo, reviewer)
			if err != nil {
				fmt.Printf("Error fetching email for %s in repo %s: %v\n", reviewer, repo, err)
				email = "No email found"
			}

			writer.Write([]string{parts[4], prLink, reviewer, email})
		}
	}

	fmt.Println("CSV file 'reviewers_with_emails.csv' has been created with PR links, reviewers, and their emails if available.")
}
