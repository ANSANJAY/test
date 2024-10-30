package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"io/ioutil"
	"encoding/json"
)

type Review struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
}

func main() {
	// Open the input CSV file
	inputFile, err := os.Open("input.csv")
	if err != nil {
		log.Fatalf("failed to open input CSV file: %v", err)
	}
	defer inputFile.Close()

	// Create the output CSV file
	outputFile, err := os.Create("output.csv")
	if err != nil {
		log.Fatalf("failed to create output CSV file: %v", err)
	}
	defer outputFile.Close()

	// Create CSV readers and writers
	reader := csv.NewReader(inputFile)
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write headers to the output CSV
	writer.Write([]string{"Pull Request Link", "Reviewers"})

	// Read the input CSV line by line
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("failed to read input CSV file: %v", err)
	}

	for _, record := range records {
		prLink := record[0]
		reviewers, err := getReviewers(prLink)
		if err != nil {
			log.Printf("failed to get reviewers for %s: %v", prLink, err)
			continue
		}

		// Join reviewers and write to output CSV
		reviewersString := strings.Join(reviewers, ", ")
		writer.Write([]string{prLink, reviewersString})
		fmt.Printf("Processed PR link: %s with reviewers: %s\n", prLink, reviewersString)
	}
}

// getReviewers retrieves the reviewers for a given pull request link
func getReviewers(prLink string) ([]string, error) {
	// Parse the pull request link
	parts := strings.Split(prLink, "/")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid pull request link format")
	}

	owner := "abc"  // Always "abc" as per user’s specification
	repo := parts[4]
	pullNumber := parts[6]

	// GitHub API URL for pull request reviews
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%s/reviews", owner, repo, pullNumber)

	// Create a new HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch reviews: %s", resp.Status)
	}

	// Parse the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var reviews []Review
	if err := json.Unmarshal(body, &reviews); err != nil {
		return nil, err
	}

	// Extract reviewer usernames
	reviewerSet := make(map[string]struct{})
	for _, review := range reviews {
		if _, exists := reviewerSet[review.User.Login]; !exists {
			reviewerSet[review.User.Login] = struct{}{}
		}
	}

	// Convert set to slice
	var reviewers []string
	for reviewer := range reviewerSet {
		reviewers = append(reviewers, reviewer)
	}

	return reviewers, nil
}
