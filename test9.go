package main

import (
    "encoding/base64"
    "encoding/csv"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strings"
    "sync"
)

// Struct to store API response content
type Content struct {
    Content string `json:"content"`
}

// Struct to store repo results
type RepoResult struct {
    RepoName  string
    IDOrError string
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run fetch_repo_data.go <input_csv>")
        return
    }

    inputCSV := os.Args[1]
    outputCSV := "output_go.csv"

    // Open input CSV
    file, err := os.Open(inputCSV)
    if err != nil {
        log.Fatalf("Failed to open input CSV: %v", err)
    }
    defer file.Close()

    // Create output CSV
    outFile, err := os.Create(outputCSV)
    if err != nil {
        log.Fatalf("Failed to create output CSV: %v", err)
    }
    defer outFile.Close()

    writer := csv.NewWriter(outFile)
    defer writer.Flush()

    // Write the header row
    writer.Write([]string{"repo_name", "id_or_error"})

    // Use a wait group to manage goroutines
    var wg sync.WaitGroup

    // Channel to collect results
    resultChan := make(chan RepoResult, 100)

    // Goroutine to write results to the output CSV
    go func() {
        for result := range resultChan {
            writer.Write([]string{result.RepoName, result.IDOrError})
        }
    }()

    // Read input CSV and start a goroutine for each repo
    reader := csv.NewReader(file)
    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatalf("Error reading input CSV: %v", err)
        }

        repoName := strings.TrimSpace(record[0])
        if repoName == "repo_name" || repoName == "" {
            continue // Skip header or empty rows
        }

        wg.Add(1)
        go func(repo string) {
            defer wg.Done()
            resultChan <- fetchRepoData(repo)
        }(repoName)
    }

    // Wait for all goroutines to complete
    wg.Wait()
    close(resultChan)

    fmt.Println("Script completed. Output saved to", outputCSV)
}

// fetchRepoData makes the API call and returns the result
func fetchRepoData(repo string) RepoResult {
    url := fmt.Sprintf("https://api.github.com/repos/amex-eng/%s/contents/.amex/buildblocks.yaml", repo)
    
    fmt.Printf("DEBUG: Fetching data for repo: %s\n", repo)
    fmt.Printf("DEBUG: API URL: %s\n", url)

    resp, err := http.Get(url)
    if err != nil {
        log.Printf("ERROR: Failed to make API request for repo %s: %v\n", repo, err)
        return RepoResult{RepoName: repo, IDOrError: fmt.Sprintf("Error: %v", err)}
    }
    defer resp.Body.Close()

    // Log the status code and check for 404s or other issues
    fmt.Printf("DEBUG: Response status for %s: %s\n", repo, resp.Status)

    if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body) // Read the full response body for debugging
        fmt.Printf("DEBUG: Response body for %s: %s\n", repo, string(body))
        return RepoResult{RepoName: repo, IDOrError: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))}
    }

    var content Content
    if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
        log.Printf("ERROR: Failed to decode JSON for repo %s: %v\n", repo, err)
        return RepoResult{RepoName: repo, IDOrError: fmt.Sprintf("Error decoding JSON: %v", err)}
    }

    // Decode the content from base64
    decoded, err := base64.StdEncoding.DecodeString(content.Content)
    if err != nil {
        log.Printf("ERROR: Failed to decode content for repo %s: %v\n", repo, err)
        return RepoResult{RepoName: repo, IDOrError: fmt.Sprintf("Error decoding content: %v", err)}
    }

    // Extract the ID from the decoded content
    id := extractID(string(decoded))
    if id == "" {
        log.Printf("WARNING: 'id' not found for repo %s\n", repo)
        return RepoResult{RepoName: repo, IDOrError: "Error: 'id' not found"}
    }

    fmt.Printf("DEBUG: Found ID for %s: %s\n", repo, id)
    return RepoResult{RepoName: repo, IDOrError: id}
}

// extractID extracts the 'id' from the decoded content
func extractID(content string) string {
    for _, line := range strings.Split(content, "\n") {
        if strings.HasPrefix(line, "id:") {
            parts := strings.Split(line, "'")
            if len(parts) > 1 {
                return parts[1]
            }
        }
    }
    return ""
}
