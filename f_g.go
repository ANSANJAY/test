package main

import (
    "encoding/csv"
    "fmt"
    "os"
)

func main() {
    // Open the input CSV file
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

    // Open the output CSV file
    outputFile, err := os.Create("output_messages.csv")
    if err != nil {
        fmt.Println("Error creating output file:", err)
        return
    }
    defer outputFile.Close()

    // Create a CSV writer
    writer := csv.NewWriter(outputFile)
    defer writer.Flush()

    // Write the header to the output CSV
    header := []string{"Repo Name", "Contributor", "Email", "Message"}
    writer.Write(header)

    // Loop through each record and write an individual message for each contributor
    for _, row := range records[1:] { // Skip the header row
        repoName := row[0]
        repoLink := row[1]
        contributorName := row[2] // Contributor name without "@"
        contributorEmail := row[3]

        // Template message without double quotes and "@" in the message body
        message := fmt.Sprintf(`Hi %s,

On behalf of the ABC team, we have raised a PR named abc for the integration of coverage software in your GitHub repository: %s (%s).

We would appreciate your help in reviewing and merging this PR.

For more details, please refer to the documentation here: [ABC documentation link].
You can also reach us on Slack at #ABC-slack-channel.

Thank you!

Best regards,
ABC Team`, contributorName, repoName, repoLink)

        // Write each row to the output CSV
        rowOutput := []string{repoName, "@" + contributorName, contributorEmail, message}
        writer.Write(rowOutput)
    }

    fmt.Println("Individual messages have been written to output_messages.csv")
}
