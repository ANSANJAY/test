package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	fileName = "xyz"   // The file name you're looking for
	textFile = "repos.txt" // Text file with repository names
)

func main() {
	// Open the text file
	file, err := os.Open(textFile)
	if err != nil {
		fmt.Println("Error opening text file:", err)
		return
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Iterate over each line in the text file
	for scanner.Scan() {
		repoName := strings.TrimSpace(scanner.Text()) // Read the repo name from the text file
		if repoName == "" {
			continue // Skip empty lines
		}

		fmt.Printf("Checking file '%s' in repository '%s'...\n", fileName, repoName)

		// Construct the `gh` command to list the root contents of the repository
		cmd := exec.Command("gh", "repo", "view", repoName, "--json", "files", "--jq", ".files[].path")

		// Run the command and capture the output
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Error executing gh command:", err)
			continue
		}

		// Check if the file exists in the root directory
		if strings.Contains(string(output), fileName) {
			fmt.Printf("File '%s' exists in the root of the repository '%s'.\n", fileName, repoName)
		} else {
			fmt.Printf("File '%s' does not exist in the root of the repository '%s'.\n", fileName, repoName)
		}
	}

	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading text file:", err)
	}
}