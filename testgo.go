package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Define the owner prefix as amex-eng/
const ownerPrefix = "amex-eng/"

// Function to get pom.xml content using the gh CLI
func getPomXML(repoName string) (string, error) {
	// Append owner prefix to the repository name
	fullRepoName := ownerPrefix + repoName

	cmd := exec.Command("gh", "api", fmt.Sprintf("repos/%s/contents/pom.xml", fullRepoName), "--jq", ".content")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	// Decode base64 content
	decodedContent, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(out.String(), "\n", ""))
	if err != nil {
		return "", err
	}

	return string(decodedContent), nil
}

// Function to check if pom.xml contains the Jacoco plugin
func hasJacocoPlugin(pomContent string) bool {
	return strings.Contains(pomContent, "jacoco-maven-plugin")
}

func main() {
	// Open the input CSV file with repository names
	inputFile, err := os.Open("repos.csv")
	if err != nil {
		log.Fatalf("Failed to open file: %s", err)
	}
	defer inputFile.Close()

	// Read the CSV file
	csvReader := csv.NewReader(inputFile)
	repoNames, err := csvReader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV file: %s", err)
	}

	// Open the output CSV file for writing repositories without Jacoco plugin
	outputFile, err := os.Create("repos_without_jacoco.csv")
	if err != nil {
		log.Fatalf("Failed to create output file: %s", err)
	}
	defer outputFile.Close()

	csvWriter := csv.NewWriter(outputFile)
	defer csvWriter.Flush()

	// Process each repository
	for _, repoRow := range repoNames {
		repoName := repoRow[0]

		// Fetch pom.xml content using gh CLI
		pomContent, err := getPomXML(repoName)
		if err != nil {
			fmt.Printf("Error fetching pom.xml for repo %s: %s\n", repoName, err)
			continue
		}

		// Check if the repository has Jacoco plugin
		if !hasJacocoPlugin(pomContent) {
			// Write the repo name (with owner prefix) to the output CSV if Jacoco plugin is not found
			err := csvWriter.Write([]string{ownerPrefix + repoName})
			if err != nil {
				log.Fatalf("Failed to write to CSV: %s", err)
			}
		}
	}

	fmt.Println("Processing completed. Check repos_without_jacoco.csv for results.")
}