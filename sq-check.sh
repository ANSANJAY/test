#!/bin/bash

# Constants
OWNER="your-owner"            # Replace with the GitHub owner (username or organization)
FILE_NAME="xyz"               # The file you are looking for
REPOS_FILE="repos.txt"        # File containing repository names (one per line)
OUTPUT_FILE="output.csv"      # CSV file to store the output

# Check if repos file exists
if [ ! -f "$REPOS_FILE" ]; then
    echo "Repository list file '$REPOS_FILE' not found!"
    exit 1
fi

# Create or overwrite the output CSV file with headers
echo "Repository,Status" > "$OUTPUT_FILE"

# Loop through each repository name in the repos file
while IFS= read -r REPO_NAME || [[ -n "$REPO_NAME" ]]; do
    REPO_NAME=$(echo "$REPO_NAME" | xargs) # Trim any leading/trailing spaces

    if [ -z "$REPO_NAME" ]; then
        continue  # Skip empty lines
    fi

    FULL_REPO_NAME="$OWNER/$REPO_NAME"

    echo "Checking file '$FILE_NAME' in repository '$FULL_REPO_NAME'..."

    # Use gh api to check if the file exists in the repository
    RESPONSE=$(gh api "repos/$FULL_REPO_NAME/contents/$FILE_NAME" 2>&1)

    # Check if the file exists or not
    if echo "$RESPONSE" | grep -q '"name":'; then
        echo "File '$FILE_NAME' exists in the repository '$FULL_REPO_NAME'."
        echo "$FULL_REPO_NAME,found" >> "$OUTPUT_FILE"
    elif echo "$RESPONSE" | grep -q '404 Not Found'; then
        echo "File '$FILE_NAME' does not exist in the repository '$FULL_REPO_NAME'."
        echo "$FULL_REPO_NAME,not found" >> "$OUTPUT_FILE"
    else
        echo "Error checking file in repository '$FULL_REPO_NAME':"
        echo "$RESPONSE"
        echo "$FULL_REPO_NAME,error" >> "$OUTPUT_FILE"
    fi
done < "$REPOS_FILE"

echo "Check completed. Results saved to $OUTPUT_FILE."