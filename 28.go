package main

import (
	"encoding/csv"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Struct for Slack message
type SlackMessage struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// Function to send Slack notification
func sendSlackNotification(slackToken, channel, message string) error {
	// Create Slack message payload
	msg := SlackMessage{
		Channel: channel,
		Text:    message,
	}

	// Convert message struct to JSON
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Prepare the HTTP request to Slack API
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+slackToken)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	// Check if the response from Slack indicates an error
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error from Slack API: %s", body)
	}

	fmt.Println("Notification sent successfully!")
	return nil
}

func main() {
	// Slack bot token (replace with your own token)
	slackToken := "xoxb-your-slack-bot-token"

	// Open the CSV file containing repo details and contributor emails
	file, err := os.Open("contributors_output.csv")
	if err != nil {
		log.Fatalf("Could not open CSV file: %v", err)
	}
	defer file.Close()

	// Read the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Could not read CSV file: %v", err)
	}

	// Iterate over each row in the CSV (assuming repo, repo link, contributor name, contributor email)
	for i, record := range records {
		// Skip header row
		if i == 0 {
			continue
		}

		repoName := record[0]
		repoLink := record[1]
		contributorEmail := record[3] // Assuming email is in the 4th column

		// Message to be sent to Slack
		message := fmt.Sprintf("Hi! A PR is open for repository: %s (%s). Please review and merge it when possible!", repoName, repoLink)

		// Send Slack notification (using contributor's email as the channel)
		err := sendSlackNotification(slackToken, contributorEmail, message)
		if err != nil {
			fmt.Printf("Error sending notification to %s: %v\n", contributorEmail, err)
		}
	}
}
