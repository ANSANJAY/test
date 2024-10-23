package main

import (
	"fmt"
	"log"

	"github.com/slack-go/slack"
)

func main() {
	// Replace with your Slack token and the recipient's Slack User ID
	token := "your-slack-token"
	recipientUserID := "recipient-user-id"

	// Initialize the Slack client
	client := slack.New(token)

	// Message you want to send
	messageText := "Hello! This is a personal message from the Go program."

	// Send the message as a direct message to the user
	channel, timestamp, err := client.PostMessage(
		recipientUserID,
		slack.MsgOptionText(messageText, false),
	)

	if err != nil {
		log.Fatalf("Failed to send message: %s", err)
	}

	fmt.Printf("Message successfully sent to channel %s at %s", channel, timestamp)
}
