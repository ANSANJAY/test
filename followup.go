package main

import (
    "encoding/csv"
    "fmt"
    "os"
)

func main() {
    // Open the CSV file
    file, err := os.Open("contributors.csv")
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer file.Close()

    // Create a CSV reader
    reader := csv.NewReader(file)
    
    // Read all records
    records, err := reader.ReadAll()
    if err != nil {
        fmt.Println("Error reading CSV:", err)
        return
    }

    // Map to store contributors by repository
    repoContributors := make(map[string][]string)
    repoLinks := make(map[string]string)

    // Populate the map with contributors and links grouped by repository
    for _, row := range records[1:] { // Skip the header row
        repoName := row[0]
        repoLink := row[1]
        contributorName := row[2]

        repoContributors[repoName] = append(repoContributors[repoName], contributorName)
        repoLinks[repoName] = repoLink
    }

    // Generate and print a single message per repository
    for repoName, contributors := range repoContributors {
        repoLink := repoLinks[repoName]
        contributorList := ""

        for _, contributor := range contributors {
            contributorList += contributor + ", "
        }

        // Trim the trailing comma and space
        contributorList = contributorList[:len(contributorList)-2]

        // Template message
        message := fmt.Sprintf(`Hi %s,

On behalf of the ABC team, we have raised a PR named "abc" for the integration of coverage software in your GitHub repository: %s (%s).

We would appreciate your help in reviewing and merging this PR.

For more details, please refer to the documentation here: [ABC documentation link].
You can also reach us on Slack at #ABC-slack-channel.

Thank you!

Best regards,
ABC Team`, contributorList, repoName, repoLink)

        // Print or save the message
        fmt.Println(message)
        fmt.Println("------")
    }
}
