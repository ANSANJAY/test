package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	owner       = "your-owner"      // Replace with the owner of the repository
	repo        = "your-repo"       // Replace with the repository name
	fileName    = "xyz"             // The file name you're looking for
	githubToken = "your-token"      // Replace with your GitHub token
)

type GitHubContent struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}

func main() {
	// GitHub API to list contents of the root directory of the repository
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/", owner, repo)

	// Create a new request with the GitHub token for authentication
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}

	// Set the Authorization header with the GitHub token
	req.Header.Set("Authorization", "Bearer "+githubToken)

	// Create a new HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Check if the response status is 200 (OK)
	if resp.StatusCode == 200 {
		// Read the body of the response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			os.Exit(1)
		}

		// Parse the JSON response
		var contents []GitHubContent
		if err := json.Unmarshal(body, &contents); err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			os.Exit(1)
		}

		// Search for the file in the root directory
		fileExists := false
		for _, content := range contents {
			if content.Name == fileName && content.Type == "file" {
				fileExists = true
				break
			}
		}

		// Print the result based on whether the file was found
		if fileExists {
			fmt.Printf("File '%s' exists in the root of the repository.\n", fileName)
		} else {
			fmt.Printf("File '%s' does not exist in the root of the repository.\n", fileName)
		}
	} else {
		fmt.Printf("Failed to list contents of the repository. HTTP status code: %d\n", resp.StatusCode)
	}
}