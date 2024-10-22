package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	// Open the input file
	inputFile, err := os.Open("input.txt")
	if err != nil {
		fmt.Println("Error opening input file:", err)
		return
	}
	defer inputFile.Close()

	// Open the output CSV file
	outputFile, err := os.Create("output.csv")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	// Create a CSV writer
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write the header row to the CSV
	header := []string{"repo_name", "central_id", "application_name", "application_owner", "application_manager", "application_vp1", "application_vp2"}
	writer.Write(header)

	// Regular expressions to capture the fields
	repoNameRegex := regexp.MustCompile(`repo_name\s*\|\s*(.+)`)
	centralIDRegex := regexp.MustCompile(`central_id\s*\|\s*(.+)`)
	applicationNameRegex := regexp.MustCompile(`application_name\s*\|\s*(.+)`)
	applicationOwnerRegex := regexp.MustCompile(`application_owner\s*\|\s*(.+)`)
	applicationManagerRegex := regexp.MustCompile(`application_manager\s*\|\s*(.+)`)
	applicationVP1Regex := regexp.MustCompile(`application_vp1\s*\|\s*(.+)`)
	applicationVP2Regex := regexp.MustCompile(`application_vp2\s*\|\s*(.+)`)

	// Variables to store the current record fields
	var repoName, centralID, applicationName, applicationOwner, applicationManager, applicationVP1, applicationVP2 string

	// Scan through the input file line by line
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for each field in the current line using regex
		if matches := repoNameRegex.FindStringSubmatch(line); matches != nil {
			repoName = strings.TrimSpace(matches[1])
		} else if matches := centralIDRegex.FindStringSubmatch(line); matches != nil {
			centralID = strings.TrimSpace(matches[1])
		} else if matches := applicationNameRegex.FindStringSubmatch(line); matches != nil {
			applicationName = strings.TrimSpace(matches[1])
		} else if matches := applicationOwnerRegex.FindStringSubmatch(line); matches != nil {
			applicationOwner = strings.TrimSpace(matches[1])
		} else if matches := applicationManagerRegex.FindStringSubmatch(line); matches != nil {
			applicationManager = strings.TrimSpace(matches[1])
		} else if matches := applicationVP1Regex.FindStringSubmatch(line); matches != nil {
			applicationVP1 = strings.TrimSpace(matches[1])
		} else if matches := applicationVP2Regex.FindStringSubmatch(line); matches != nil {
			applicationVP2 = strings.TrimSpace(matches[1])
		}

		// If we've gathered all fields for a record, write to the CSV
		if repoName != "" && centralID != "" && applicationName != "" && applicationOwner != "" && applicationManager != "" && applicationVP1 != "" && applicationVP2 != "" {
			record := []string{repoName, centralID, applicationName, applicationOwner, applicationManager, applicationVP1, applicationVP2}
			writer.Write(record)

			// Reset the fields for the next record
			repoName, centralID, applicationName, applicationOwner, applicationManager, applicationVP1, applicationVP2 = "", "", "", "", "", "", ""
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input file:", err)
	}
}
