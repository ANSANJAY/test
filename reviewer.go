package main

import (
    "encoding/csv"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "net/url"
    "os"
    "regexp"
    "strings"
)

const (
    org = "amex-eng" // Organization name
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
    writer.Write([]string{"link", "reviewers"})

    for {
        record, err := reader.Read()
        if err != nil {
            break
        }
        
        link := record[0]
        repo, pullNumber, err := extractRepoAndPull(link)
        if err != nil {
            log.Printf("Failed to extract repo/pull from link %s: %v", link, err)
            writer.Write([]string{link, "Error extracting repo/pull"})
            continue
        }

        reviewers, err := fetchReviewers(repo, pullNumber)
        if err != nil {
            log.Printf("Failed to fetch reviewers for %s/%s: %v", repo, pullNumber, err)
            writer.Write([]string{link, "Error fetching reviewers"})
            continue
        }

        writer.Write([]string{link, strings.Join(reviewers, ", ")})
    }

    fmt.Println("Reviewers fetched and written to output.csv successfully.")
}

func extractRepoAndPull(link string) (string, string, error) {
    parsedURL, err := url.Parse(link)
    if err != nil {
        return "", "", err
    }

    // Use regex to extract repo and pull number
    re := regexp.MustCompile(`/([^/]+)/pull/(\d+)`)
    matches := re.FindStringSubmatch(parsedURL.Path)
    if len(matches) < 3 {
        return "", "", fmt.Errorf("could not extract repo and pull number from link")
    }

    repo := matches[1]
    pullNumber := matches[2]
    return repo, pullNumber, nil
}

func fetchReviewers(repo, pullNumber string) ([]string, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%s/requested_reviewers", org, repo, pullNumber)

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
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
