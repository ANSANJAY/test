package main

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	owner    = "your-owner" // Replace with the owner of the repository
	repo     = "your-repo"  // Replace with the repository name
	fileName = "xyz"        // The file name you're looking for
)

func main() {
	// Construct the `gh` command to list the root contents of the repository
	cmd := exec.Command("gh", "repo", "view", fmt.Sprintf("%s/%s", owner, repo), "--json", "files", "--jq", ".files[].path")

	// Run the command and capture the output
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing gh command:", err)
		return
	}

	// Check if the file exists in the output
	if strings.Contains(string(output), fileName) {
		fmt.Printf("File '%s' exists in the root of the repository.\n", fileName)
	} else {
		fmt.Printf("File '%s' does not exist in the root of the repository.\n", fileName)
	}
}