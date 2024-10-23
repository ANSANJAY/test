package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	owner    = "your-owner" // Replace with the GitHub owner name (organization or username)
	fileName = "xyz"        // The file name you're looking for
	textFile = "repos.txt"  // Text file with repository names
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

		// Combine owner and repository name
		fullRepoName := fmt.Sprintf("%s/%s", owner, repoName)

		fmt.Printf("Checking file '%s' in repository '%s'...\n", fileName, fullRepoName)

		// Construct the gh command to list the root contents of the repository
		cmd := exec.Command("gh", "repo", "view", fullRepoName, "--json", "files", "--jq", ".files[].path")

		// Capture the standard output and error
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error executing gh command for repo '%s': %s\n", fullRepoName, err.Error())
			fmt.Printf("Command output: %s\n", string(output))
			continue
		}

		// Check if the file exists in the root directory
		if strings.Contains(string(output), fileName) {
			fmt.Printf("File '%s' exists in the root of the repository '%s'.\n", fileName, fullRepoName)
		} else {
			fmt.Printf("File '%s' does not exist in the root of the repository '%s'.\n", fileName, fullRepoName)
		}
	}

	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading text file:", err)
	}
}