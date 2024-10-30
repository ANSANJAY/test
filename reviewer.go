package main

import (
    "encoding/csv"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "strings"
)

const (
    githubToken = "YOUR_GITHUB_TOKEN" // Replace with your GitHub token
    org         = "amex-eng"          // Organization name
)

func main() {
    inputFile, err := os.Open("input.csv")
    if err != nil {
        log.Fatalf("Failed to open input file: %v", err)
    }
    defer inputFile.Close()

    reader := csv.NewReader(inputFile)
    headers, err := reader.Read() // Read the header
    if err != nil {
        log.Fatalf("Failed to read CSV header: %v", err)
    }

    // Output file setup
    outputFile, err := os.Create("output.csv")
    if err != nil {
        log.Fatalf("Failed to create output file: %v", err)
    }
    defer outputFile.Close()

    writer := csv.NewWriter(outputFile)
    defer writer.Flush()

    // Write headers to output
    headers = append(headers, "reviewers")
    writer.Write(headers)

    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatalf("Error reading CSV record: %v", err)
        }

        repo := record[0]          // Assuming first column is repo
        pullNumber := record[1]    // Assuming second column is pull_number

        reviewers, err := fetchReviewers(repo, pullNumber)
        if err != nil {
            log.Printf("Failed to fetch reviewers for %s/%s#%s: %v", org, repo, pullNumber, err)
            record = append(record, "Error fetching reviewers")
        } else {
            record = append(record, strings.Join(reviewers, ", "))
        }

        writer.Write(record)
    }

    fmt.Println("Reviewers fetched and written to output.csv successfully.")
}

func fetchReviewers(repo, pullNumber string) ([]string, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%s/requested_reviewers", org, repo, pullNumber)

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", "token "+githubToken)
    req.Header.Set("Accept", "application/vnd.github.v3+json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("GitHub API responded with status: %v", resp.Status)
    }

    var result struct {
        Users []struct {
            Login string `json:"login"`
        } `json:"users"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    var reviewers []string
    for _, user := range result.Users {
        reviewers = append(reviewers, user.Login)
    }
    return reviewers, nil
}
