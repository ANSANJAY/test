package main

import (
    "encoding/csv"
    "fmt"
    "os"
)

func main() {
    // Open contributors.csv
    contributorsFile, err := os.Open("contributors.csv")
    if err != nil {
        fmt.Println("Error opening contributors file:", err)
        return
    }
    defer contributorsFile.Close()

    // Read contributors.csv
    contributorsReader := csv.NewReader(contributorsFile)
    contributorsRecords, err := contributorsReader.ReadAll()
    if err != nil {
        fmt.Println("Error reading contributors CSV:", err)
        return
    }

    // Open pull_requests.csv
    pullRequestsFile, err := os.Open("pull_requests.csv")
    if err != nil {
        fmt.Println("Error opening pull requests file:", err)
        return
    }
    defer pullRequestsFile.Close()

    // Read pull_requests.csv
    pullRequestsReader := csv.NewReader(pullRequestsFile)
    pullRequestsRecords, err := pullRequestsReader.ReadAll()
    if err != nil {
        fmt.Println("Error reading pull requests CSV:", err)
        return
    }

    // Map to store pull request links by repository name
    pullRequestsMap := make(map[string]string)
    for _, prRow := range pullRequestsRecords[1:] { // Skip header row
        if len(prRow) < 2 {
            fmt.Println("Skipping incomplete row in pull_requests.csv:", prRow)
            continue
        }
        repoName := prRow[0]
        pullRequestLink := prRow[1]
        pullRequestsMap[repoName] = pullRequestLink
    }

    // Create output CSV file
    outputFile, err := os.Create("output_messages.csv")
    if err != nil {
        fmt.Println("Error creating output file:", err)
        return
    }
    defer outputFile.Close()

    // Create a CSV writer for output
    writer := csv.NewWriter(outputFile)
    defer writer.Flush()

    // Write header to output CSV
    header := []string{"Repo Name", "Contributor", "Email", "Pull Request", "Message"}
    writer.Write(header)

    // Loop through contributors.csv records
    for _, row := range contributorsRecords[1:] { // Skip header row
        if len(row) < 4 {
            fmt.Println("Skipping incomplete row in contributors.csv:", row)
            continue
        }
        repoName := row[0]
        repoLink := row[1]
        contributorName := row[2]
        contributorEmail := row[3]

        // Find the pull request link for this repository
        pullRequestLink := pullRequestsMap[repoName]
        if pullRequestLink == "" {
            pullRequestLink = "No PR Link Available"
        }

        // Template message without contributor's name
        message := fmt.Sprintf(`Hi,

On behalf of the ABC team, we have raised a PR named abc for the integration of coverage software in your GitHub repository: %s (%s).

Pull Request Link: %s

We would appreciate your help in reviewing and merging this PR.

For more details, please refer to the documentation here: [ABC documentation link].
You can also reach us on Slack at #ABC-slack-channel.

Thank you!

Best regards,
ABC Team`, repoName, repoLink, pullRequestLink)

        // Write row to output CSV
        rowOutput := []string{repoName, "@" + contributorName, contributorEmail, pullRequestLink, message}
        writer.Write(rowOutput)
    }

    fmt.Println("Individual messages with pull requests have been written to output_messages.csv")
}
