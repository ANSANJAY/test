package main

import (
    "encoding/csv"
    "fmt"
    "os"
    "strings"
)

func main() {
    // Open the input CSV file with contributors
    file, err := os.Open("contributors.csv")
    if err != nil {
        fmt.Println("Error opening contributors file:", err)
        return
    }
    defer file.Close()

    // Create a CSV reader for the contributors file
    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        fmt.Println("Error reading contributors CSV:", err)
        return
    }

    // Open the pull requests CSV file
    prFile, err := os.Open("pull_requests.csv")
    if err != nil {
        fmt.Println("Error opening pull requests file:", err)
        return
    }
    defer prFile.Close()

    // Create a CSV reader for the pull requests file
    prReader := csv.NewReader(prFile)
    prRecords, err := prReader.ReadAll()
    if err != nil {
        fmt.Println("Error reading pull requests CSV:", err)
        return
    }

    // Map to store pull request links by repository name
    prLinks := make(map[string]string)
    for _, prRow := range prRecords[1:] { // Skip header row
        if len(prRow) < 1 {
            fmt.Println("Skipping incomplete row in pull_requests.csv:", prRow)
            continue
        }
        pullLink := prRow[0]
        
        // Extract repository name from pull request link
        parts := strings.Split(pullLink, "/")
        if len(parts) < 5 {
            fmt.Println("Skipping malformed pull request link:", pullLink)
            continue
        }
        repoName := parts[4] // Extract repo name from URL path
        prLinks[repoName] = pullLink
    }

    // Open the output CSV file
    outputFile, err := os.Create("output_messages.csv")
    if err != nil {
        fmt.Println("Error creating output file:", err)
        return
    }
    defer outputFile.Close()

    // Create a CSV writer for the output file
    writer := csv.NewWriter(outputFile)
    defer writer.Flush()

    // Write the header to the output CSV
    header := []string{"Repo Name", "Contributor", "Email", "Pull Request", "Message"}
    writer.Write(header)

    // Loop through each contributor record
    for _, row := range records[1:] { // Skip header row
        if len(row) < 4 {
            fmt.Println("Skipping incomplete row in contributors.csv:", row)
            continue
        }
        repoName := row[0]
        repoLink := row[1] // Directly use the repo link from contributors.csv
        contributorName := row[2]
        contributorEmail := row[3]

        // Retrieve the pull request link for this repository
        pullRequestLink, exists := prLinks[repoName]
        if !exists {
            pullRequestLink = "No PR Link Available"
        }

        // Message that uses the repo link directly from contributors.csv
        message := fmt.Sprintf(`Hi,

On behalf of the ABC team, we have raised a PR named abc for the integration of coverage software in your GitHub repository: %s.

Pull Request Link: %s

We would appreciate your help in reviewing and merging this PR.

For more details, please refer to the documentation here: [ABC documentation link].
You can also reach us on Slack at #ABC-slack-channel.

Thank you!

Best regards,
ABC Team`, repoLink, pullRequestLink)

        // Write each row to the output CSV
        rowOutput := []string{repoName, "@" + contributorName, contributorEmail, pullRequestLink, message}
        writer.Write(rowOutput)
    }

    fmt.Println("Individual messages with pull requests have been written to output_messages.csv")
}
